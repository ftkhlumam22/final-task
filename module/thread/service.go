package threadmodule

func NewThreadService(
	threadReadRepository ThreadReadRepository,
	threadListCacheRepository ThreadListCacheRepository,
	threadDetailCacheRepository ThreadDetailCacheRepository,
	threadEventPublisher ThreadEventPublisher,
	threadRPCPublisher ThreadRPCPublisher,
) *ThreadService {
	return &ThreadService{
		threadReadRepository:        threadReadRepository,
		threadListCacheRepository:   threadListCacheRepository,
		threadDetailCacheRepository: threadDetailCacheRepository,
		threadEventPublisher:        threadEventPublisher,
		threadRPCPublisher:          threadRPCPublisher,
	}
}
