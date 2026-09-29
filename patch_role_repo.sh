#!/bin/bash
cat << 'INNER_EOF' > /tmp/role_repo_patch.go
	FindTrashedAll(page, perPage int, search, sort, order string) ([]models.Role, int64, error)
	Restore(id uuid.UUID) error
	ForceDelete(id uuid.UUID) error
INNER_EOF
sed -i '' '/Delete(id uuid.UUID) error/r /tmp/role_repo_patch.go' pkg/repositories/role_repository.go
