.PHONY: gen-mocks
gen-mocks:
	mockgen -destination=internal/repositories/mocks/mock_identity.go -package=mocks gophkeeper/internal/repositories/identity Identifier && \
	mockgen -destination=internal/repositories/mocks/mock_server_storage.go -package=mocks gophkeeper/internal/server/storage IEncryptedServerStorage