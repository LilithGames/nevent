.PHONY: build
build:
	@go build github.com/LilithGames/nevent/protoc-gen-nevent/...

.PHONY: clean
clean:
	@rm -f *.exe
	@rm -rf proto/pb


.PHONY: proto-event
proto-event:
	@protoc -I=. --go_out=paths=source_relative:. proto/nevent.proto

.PHONY: proto
proto: clean build
	@protoc -I=. --go_out=paths=source_relative:. --nevent_out=paths=source_relative:. testdata/proto/test.proto

.PHONY: tag
tag:
	@git tag $$(svu next)
