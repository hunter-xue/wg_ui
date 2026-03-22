.PHONY: build linux deploy test clean

BUILD_DIR := build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/wg_ui .

linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/wg_ui_linux .

deploy: linux
ifndef HOST
	$(error HOST is required, e.g. make deploy HOST=1.2.3.4)
endif
	scp $(BUILD_DIR)/wg_ui_linux root@$(HOST):/root/wg_ui

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)
