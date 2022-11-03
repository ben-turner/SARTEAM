
GOSRCS := $(shell find . -name '*.go')

.PHONY: clean run buildandrun web/dist

web/dist:
	cd web && npm run build

bin/sarteam: $(GOSRCS)
	go build -o bin/sarteam ./cmd/sarteam

run: bin/sarteam
	SARTEAM_WEBDIR=web/dist bin/sarteam

buildandrun: web/dist bin/sarteam run

clean:
	rm -rf bin
	rm -rf dist
