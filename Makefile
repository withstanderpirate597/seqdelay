.PHONY: test bench vet build docs docs-dev docs-deploy clean

# === Go ===

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./... -race -timeout 120s

test-short:
	go test ./... -short -timeout 30s

test-stress:
	go test -run TestStress -v -timeout 120s

bench:
	go test -bench=BenchmarkTimeWheel -benchmem -count=1 -timeout 60s

bench-all:
	go test -bench=. -benchmem -count=1 -timeout 120s

# === Docs ===

docs:
	cd docs && pnpm install && pnpm clear && pnpm build

docs-dev:
	cd docs && pnpm install && pnpm clear && pnpm build && npx docusaurus serve

docs-deploy:
	@cd docs && pnpm install && pnpm clear && pnpm build
	@npx wrangler whoami >/dev/null 2>&1 || npx wrangler login
	cd docs && npx wrangler pages deploy build --project-name=seqdelay-docs

# === Clean ===

clean:
	rm -rf docs/build docs/.docusaurus docs/node_modules
