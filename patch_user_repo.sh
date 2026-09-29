#!/bin/bash
cat << 'INNER_EOF' > /tmp/user_repo_patch.go
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.User, int64, error)
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
INNER_EOF
sed -i '' '/Delete(id uuid.UUID) error/r /tmp/user_repo_patch.go' pkg/repositories/user_repository.go
