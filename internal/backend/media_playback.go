package backend

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func (a *App) handlePlaybackInfo(w http.ResponseWriter, r *http.Request) {
	resolved := a.resolveRouteID(r.PathValue("itemId"))
	if resolved == nil {
		if a.Logger != nil {
			a.Logger.Warnf("PlaybackInfo: itemId=%s not found in mappings", r.PathValue("itemId"))
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Item not found"})
		return
	}
	reqCtx := requestContextFrom(r.Context())
	if !a.requireServerAccess(w, r, resolved) {
		return
	}
	// 并发限制校验
	var serverMaxConcurrent int
	var serverName string
	for _, u := range a.ConfigStore.Snapshot().Upstream {
		if u.ID == resolved.ServerID {
			serverMaxConcurrent = u.MaxConcurrent
			serverName = u.Name
			break
		}
	}

	isAdmin := reqCtx != nil && reqCtx.ProxyUser != nil && reqCtx.ProxyUser.Role == "admin"
	limiterUserID := ""
	username := ""
	if reqCtx != nil && reqCtx.ProxyUser != nil {
		limiterUserID = reqCtx.ProxyUser.UserID
		username = reqCtx.ProxyUser.Username
	}

	clientName, clientVer, deviceName, deviceID := extractClientDeviceInfo(r)
	sessionKey := buildSessionKey(limiterUserID, deviceID)

	if a.PlaybackLimiter != nil && limiterUserID != "" {
		if !isAdmin {
			userMaxConcurrent := 1
			if a.UserStore != nil {
				if u := a.UserStore.Get(limiterUserID); u != nil {
					userMaxConcurrent = u.MaxConcurrent
				}
			}
			allowed, reason := a.PlaybackLimiter.TryStartSession(&ActiveSession{
				SessionKey:    sessionKey,
				UserID:        limiterUserID,
				Username:      username,
				Client:        clientName,
				ClientVersion: clientVer,
				DeviceName:    deviceName,
				DeviceID:      deviceID,
				IP:            clientIP(r, a.ConfigStore.Snapshot().Server.TrustProxy),
				ServerID:      resolved.ServerID,
				ServerName:    serverName,
				ItemID:        r.PathValue("itemId"),
				PlayMethod:    "DirectStream",
			}, serverMaxConcurrent, userMaxConcurrent)

			if !allowed {
				writeJSON(w, http.StatusTooManyRequests, map[string]any{"message": reason})
				return
			}
		} else {
			// 管理员会话：仅登记会话信息，不设并发上限
			a.PlaybackLimiter.RecordAdminSession(&ActiveSession{
				SessionKey:    sessionKey,
				UserID:        limiterUserID,
				Username:      username,
				Client:        clientName,
				ClientVersion: clientVer,
				DeviceName:    deviceName,
				DeviceID:      deviceID,
				IP:            clientIP(r, a.ConfigStore.Snapshot().Server.TrustProxy),
				ServerID:      resolved.ServerID,
				ServerName:    serverName,
				ItemID:        r.PathValue("itemId"),
				PlayMethod:    "DirectStream",
			})
		}
	}
	instances := a.collectAllowedInstances(reqCtx, resolved)
	if a.Logger != nil {
		a.Logger.Debugf("PlaybackInfo: itemId=%s → server=[%s] originalId=%s, instances=%d",
			r.PathValue("itemId"), resolved.Client.Name, resolved.OriginalID, len(instances))
	}
	query := cloneValues(r.URL.Query())
	body := map[string]any{}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if mediaSourceID, ok := body["MediaSourceId"].(string); ok {
		if msResolved := a.IDStore.ResolveVirtualID(mediaSourceID); msResolved != nil {
			body["MediaSourceId"] = msResolved.OriginalID
		}
	}
	if mediaSourceID := query.Get("MediaSourceId"); mediaSourceID != "" {
		if msResolved := a.IDStore.ResolveVirtualID(mediaSourceID); msResolved != nil {
			query.Set("MediaSourceId", msResolved.OriginalID)
		}
	}

	// Remove proxy token from query before forwarding to upstream
	query.Del("api_key")
	query.Del("ApiKey")

	var base map[string]any
	allMediaSources := []map[string]any{}
	for _, inst := range instances {
		// The client only ever holds EIO's virtual user ID, and Emby prefers a UserId in the
		// query or body over the one the request was authenticated as. Forwarding the virtual
		// ID made the upstream look up a user that does not exist there and answer 500
		// (NullReferenceException), which surfaced as a 502 to the client.
		instQuery := cloneValues(query)
		instQuery.Set("UserId", inst.Client.clientUserID())
		instBody := deepCloneMap(body)
		instBody["UserId"] = inst.Client.clientUserID()
		payload, err := inst.Client.RequestJSON(r.Context(), requestContextFrom(r.Context()), a.Identity, r.Method, "/Items/"+inst.OriginalID+"/PlaybackInfo", instQuery, instBody)
		if err != nil {
			if a.Logger != nil {
				a.Logger.Warnf("PlaybackInfo: server=[%s] originalId=%s failed: %s", inst.Client.Name, inst.OriginalID, redactURLInError(err))
			}
			continue
		}
		data, ok := payload.(map[string]any)
		if !ok {
			continue
		}
		if base == nil {
			base = deepCloneMap(data)
		}
		for _, raw := range asItems(map[string]any{"Items": data["MediaSources"]}) {
			mediaSource := deepCloneMap(raw)
			originalMSID, _ := mediaSource["Id"].(string)
			virtualMSID := originalMSID
			if originalMSID != "" {
				virtualMSID = a.IDStore.GetOrCreateVirtualID(originalMSID, inst.ServerID)
				mediaSource["Id"] = virtualMSID
			}
			if directURL, ok := mediaSource["DirectStreamUrl"].(string); ok && directURL != "" {
				// Extract container from URL, stripping query string first
				// Node.js uses regex /\.([a-z0-9]+)(?:\?|$)/i — path.Ext doesn't stop at '?'
				cleanURL := directURL
				if qIdx := strings.IndexByte(cleanURL, '?'); qIdx >= 0 {
					cleanURL = cleanURL[:qIdx]
				}
				container := strings.TrimPrefix(path.Ext(cleanURL), ".")
				if container == "" {
					if rawContainer, ok := mediaSource["Container"].(string); ok {
						container = rawContainer
					}
				}
				if container == "" {
					container = "mp4"
				}
				isRedirect := a.streamPlaybackMode(inst.Client) == "redirect"
				if isRedirect {
					upstreamParams := url.Values{}
					if originalMSID != "" {
						upstreamParams.Set("MediaSourceId", originalMSID)
					}
					upstreamParams.Set("Static", "true")
					upstreamPath := "/Videos/" + inst.OriginalID + "/stream." + container
					reqCtx := requestContextFrom(r.Context())
					if builtURL, err := inst.Client.BuildURL(upstreamPath, upstreamParams, true, reqCtx); err == nil {
						mediaSource["DirectStreamUrl"] = builtURL
						mediaSource["Path"] = builtURL
						mediaSource["Protocol"] = "Http"
						mediaSource["IsRemote"] = true
					} else {
						proxyURL := url.Values{}
						proxyURL.Set("MediaSourceId", virtualMSID)
						proxyURL.Set("Static", "true")
						if reqCtx != nil && reqCtx.ProxyToken != "" {
							proxyURL.Set("api_key", reqCtx.ProxyToken)
						}
						mediaSource["DirectStreamUrl"] = "/Videos/" + r.PathValue("itemId") + "/stream." + container + "?" + proxyURL.Encode()
					}
				} else {
					proxyURL := url.Values{}
					proxyURL.Set("MediaSourceId", virtualMSID)
					proxyURL.Set("Static", "true")
					if reqCtx := requestContextFrom(r.Context()); reqCtx != nil && reqCtx.ProxyToken != "" {
						proxyURL.Set("api_key", reqCtx.ProxyToken)
					}
					mediaSource["DirectStreamUrl"] = "/Videos/" + r.PathValue("itemId") + "/stream." + container + "?" + proxyURL.Encode()
				}
			}
			if transcodingURL, ok := mediaSource["TranscodingUrl"].(string); ok && transcodingURL != "" {
				parsed, err := url.Parse(transcodingURL)
				if err == nil {
					if parsed.Scheme == "" {
						parsed, _ = url.Parse(inst.Client.StreamBaseURL + transcodingURL)
					}
					proxyPath := parsed.Path
					proxyPath = strings.Replace(proxyPath, "/Videos/"+inst.OriginalID+"/", "/Videos/"+r.PathValue("itemId")+"/", 1)
					proxyPath = strings.Replace(proxyPath, "/Audio/"+inst.OriginalID+"/", "/Audio/"+r.PathValue("itemId")+"/", 1)
					queryValues := parsed.Query()
					queryValues.Del("api_key")
					queryValues.Del("ApiKey")
					if originalMSID != "" {
						queryValues.Set("MediaSourceId", virtualMSID)
					}
					if reqCtx := requestContextFrom(r.Context()); reqCtx != nil && reqCtx.ProxyToken != "" {
						queryValues.Set("api_key", reqCtx.ProxyToken)
					}
					mediaSource["TranscodingUrl"] = proxyPath + "?" + queryValues.Encode()
				}
			}
			// Rewrite MediaSource.Path for Http protocol and MediaStreams[].DeliveryUrl (only when not redirect mode)
			if a.streamPlaybackMode(inst.Client) != "redirect" {
				if protocol, _ := mediaSource["Protocol"].(string); protocol == "Http" {
					if msPath, ok := mediaSource["Path"].(string); ok && msPath != "" {
						msPath = strings.ReplaceAll(msPath, inst.OriginalID, r.PathValue("itemId"))
						if originalMSID != "" && originalMSID != inst.OriginalID {
							msPath = strings.ReplaceAll(msPath, originalMSID, virtualMSID)
						}
						mediaSource["Path"] = msPath
					}
				}
			}
			if rawStreams, ok := mediaSource["MediaStreams"].([]any); ok {
				for _, rawStream := range rawStreams {
					if stream, ok := rawStream.(map[string]any); ok {
						if deliveryURL, ok := stream["DeliveryUrl"].(string); ok && deliveryURL != "" {
							deliveryURL = strings.ReplaceAll(deliveryURL, inst.OriginalID, r.PathValue("itemId"))
							if originalMSID != "" && originalMSID != inst.OriginalID {
								deliveryURL = strings.ReplaceAll(deliveryURL, originalMSID, virtualMSID)
							}
							if parsed, err := url.Parse(deliveryURL); err == nil {
								queryValues := parsed.Query()
								queryValues.Del("api_key")
								queryValues.Del("ApiKey")
								if reqCtx != nil && reqCtx.ProxyToken != "" {
									queryValues.Set("api_key", reqCtx.ProxyToken)
								}
								parsed.RawQuery = queryValues.Encode()
								deliveryURL = parsed.String()
							}
							stream["DeliveryUrl"] = deliveryURL
						}
					}
				}
			}
			allMediaSources = append(allMediaSources, mediaSource)
		}
	}
	// Record which server served this virtual item so Sessions/Playing routes back correctly
	if len(allMediaSources) > 0 {
		if virtualItemID := r.PathValue("itemId"); virtualItemID != "" {
			a.IDStore.SetActiveStream(virtualItemID, resolved.ServerID)
		}
	}
	if base == nil {
		if a.Logger != nil {
			a.Logger.Errorf("PlaybackInfo: all upstream requests failed for itemId=%s", r.PathValue("itemId"))
		}
		// Nothing will play, so release the slot TryStart reserved above. Clients retry
		// PlaybackInfo on failure, and holding the slot kept the user counted against
		// the server's capacity for the full heartbeat timeout while they could not
		// actually start a stream.
		if limiterUserID, ok := a.playbackLimiterKey(reqCtx, r, resolved.ServerID); ok {
			a.PlaybackLimiter.StopSession(sessionKey)
			a.PlaybackLimiter.Stop(limiterUserID, resolved.ServerID)
		}
		writeJSON(w, http.StatusBadGateway, map[string]any{"message": "Failed to fetch playback info from upstream"})
		return
	}
	if a.Logger != nil {
		a.Logger.Debugf("PlaybackInfo: returning %d MediaSources for itemId=%s", len(allMediaSources), r.PathValue("itemId"))
	}
	cfg := a.ConfigStore.Snapshot()
	// Rewrite top-level fields (excluding MediaSources which were already virtualised per-server above)
	delete(base, "MediaSources")
	rewriteResponseIDs(base, resolved.ServerID, a.IDStore, cfg.Server.ID, a.clientFacingUserIDFor(r))
	base["MediaSources"] = make([]any, 0, len(allMediaSources))
	for _, mediaSource := range allMediaSources {
		base["MediaSources"] = append(base["MediaSources"].([]any), mediaSource)
	}
	writeJSON(w, http.StatusOK, base)
}

func deepCloneMap(source map[string]any) map[string]any {
	encoded, _ := json.Marshal(source)
	var decoded map[string]any
	_ = json.Unmarshal(encoded, &decoded)
	return decoded
}

// resolveMediaSourceInPath resolves a virtual MediaSourceId embedded as the first path
// segment when followed by /Subtitles/ or /Attachments/ (e.g. "{msId}/Subtitles/0/Stream.srt").
func resolveMediaSourceInPath(rest string, idStore *IDStore) string {
	slash := strings.IndexByte(rest, '/')
	if slash <= 0 {
		return rest
	}
	firstSeg := rest[:slash]
	after := rest[slash:] // includes leading '/'
	if !strings.HasPrefix(after, "/Subtitles/") && !strings.HasPrefix(after, "/Attachments/") {
		return rest
	}
	if resolved := idStore.ResolveVirtualID(firstSeg); resolved != nil {
		return resolved.OriginalID + after
	}
	return rest
}
