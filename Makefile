.PHONY: gen-mocks
gen-mocks:
	mockgen -destination=internal/repositories/mocks/mock_identity.go -package=mocks github.com/abezemskiy/gophkeeper/internal/repositories/identity Identifier && \
	mockgen -destination=internal/repositories/mocks/mock_server_storage.go -package=mocks github.com/abezemskiy/gophkeeper/internal/server/storage IEncryptedServerStorage && \
	mockgen -destination=internal/repositories/mocks/mock_client_storage.go -package=mocks github.com/abezemskiy/gophkeeper/internal/client/storage IEncryptedClientStorage && \
	mockgen -destination=internal/repositories/mocks/mock_client_identity.go -package=mocks github.com/abezemskiy/gophkeeper/internal/client/identity ClientIdentifier && \
	mockgen -destination=internal/repositories/mocks/mock_client_info.go -package=mocks github.com/abezemskiy/gophkeeper/internal/client/identity IUserInfoStorage

.PHONY: build
build: build-server-client

.PHONY: build-server-agent
build-server-client: gen-proto
	go build -o cmd/client/client ./cmd/client
	go build -o cmd/server/server ./cmd/server

.PHONY: test-coverpkg
test-coverpkg:
	@INCLUDE_PACKAGES=$$(go list ./... | grep -v -E '/mocks|/protoc|/tui') && \
	go test -coverpkg=$$(echo $$INCLUDE_PACKAGES | tr ' ' ',') -coverprofile=coverage_raw.out $$INCLUDE_PACKAGES && \
	grep -v -E "github.com/abezemskiy/gophkeeper/cmd/server/main.go|github.com/abezemskiy/gophkeeper/cmd/client/main.go" coverage_raw.out > coverage.out && \
	rm coverage_raw.out