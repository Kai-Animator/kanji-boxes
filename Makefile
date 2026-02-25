.PHONY: build install test lint fmt coverage run typecheck clean

# バイナリ名
BINARY := kanji-box

# バージョン情報（gitタグまたはコミットハッシュ）
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# ビルド
build:
	go build $(LDFLAGS) -o $(BINARY) .

# テスト
test:
	go test ./...

# テスト（詳細出力）
test-v:
	go test -v ./...

# リンター実行
lint:
	golangci-lint run

# フォーマット
fmt:
	gofmt -w .
	goimports -w .

# 型チェック
typecheck:
	go build ./...

# カバレッジ
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# カバレッジ（コンソール表示）
coverage-text:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# 実行
run:
	go run .

# インストール
install:
	mkdir -p "$(HOME)/.local/bin"
	go build $(LDFLAGS) -o "$(HOME)/.local/bin/$(BINARY)" .

# クリーン
clean:
	rm -f $(BINARY) coverage.out coverage.html

# 全チェック実行
check: typecheck lint test
	@echo "All checks passed!"

# 開発用：フォーマット→型チェック→テスト
dev: fmt typecheck test
	@echo "Ready for commit!"
