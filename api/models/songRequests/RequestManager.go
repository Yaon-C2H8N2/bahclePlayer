package songRequests

type RequestManager struct {
	requests map[string]SongRequest
}

func GetRequestManager() *RequestManager {
	return &RequestManager{
		requests: make(map[string]SongRequest),
	}
}

func (rm *RequestManager) AddRequest(request SongRequest) {
	rm.requests[request.TwitchPollID] = request
}

func (rm *RequestManager) GetRequest(pollId string) SongRequest {
	var request, ok = rm.requests[pollId]
	if ok {
		return request
	}
	return SongRequest{}
}

func (rm *RequestManager) RemoveRequest(pollId string) {
	delete(rm.requests, pollId)
}

func (rm *RequestManager) GetAllRequests() map[string]SongRequest {
	return rm.requests
}

func (rm *RequestManager) ClearRequests() {
	rm.requests = make(map[string]SongRequest)
}
