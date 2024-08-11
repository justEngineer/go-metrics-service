.PHONY: build

SERVER_SOURCE_PATH=./cmd/server
AGENT_SOURCE_PATH=./cmd/agent

BUILD_VERSION_PATH=github.com/justEngineer/go-metrics-service/internal/buildversion
BUILD_VERSION=0.0.0
BUILD_DATE := $(shell date +"%Y%m%d%H%M")
BUILD_COMMIT := $(BUILD_COMMIT)

build:
	@echo "Build server is started"
	@go build -ldflags "-X '$(BUILD_VERSION_PATH).BuildVersion=$(BUILD_VERSION)' -X '$(BUILD_VERSION_PATH).BuildDate=$(BUILD_DATE)' -X '$(BUILD_VERSION_PATH).BuildCommit=$(BUILD_COMMIT)'" \
 		-o $(SERVER_SOURCE_PATH)/server $(SERVER_SOURCE_PATH)/*.go
	@echo "Build agent is started"
	@go build -ldflags "-X '$(BUILD_VERSION_PATH).BuildVersion=$(BUILD_VERSION)' -X '$(BUILD_VERSION_PATH).BuildDate=$(BUILD_DATE)' -X '$(BUILD_VERSION_PATH).BuildCommit=$(BUILD_COMMIT)'" \
	 	-o $(AGENT_SOURCE_PATH)/agent $(AGENT_SOURCE_PATH)/*.go
