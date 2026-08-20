ZENNCODE := cd tools/zenncode && go run .

# ARGS passes paths or flags through, e.g. make verify ARGS="articles/foo.md"
ARGS ?=

.PHONY: verify fix fix-dry baseline list test

verify: ## 記事中のGoサンプルコードをコンパイルし、Playgroundリンクを照合する
	$(ZENNCODE) verify $(ARGS)

fix: ## 検証を通ったスニペットを共有し、Playgroundリンクを記事に書き込む
	$(ZENNCODE) fix $(ARGS)

fix-dry: ## fix が何をするかだけ表示する(共有も書き込みもしない)
	$(ZENNCODE) fix -n $(ARGS)

baseline: ## 現時点で検証を通らないスニペットを zenncode-baseline.json に記録する
	$(ZENNCODE) baseline $(ARGS)

list: ## スニペットを1行ずつ一覧する
	$(ZENNCODE) list $(ARGS)

test: ## zenncode 自身のテスト
	cd tools/zenncode && go test ./...
