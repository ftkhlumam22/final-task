package repository

import "kaktus-consumer/repository/thread"

type ThreadRepository = thread.Repository
type ThreadTransactionRepository = thread.TransactionRepository
type ThreadRepositoryDependency = thread.Dependency

func NewThreadRepository(dependency ThreadRepositoryDependency) ThreadRepository {
	return thread.NewRepository(dependency)
}
