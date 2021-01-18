proto_clean:
	rm -rfv app/pb

proto_gen:
	mkdir app/pb
	protoc -I app/proto/ --go_out=paths=source_relative:app/pb app/proto/*.proto

build:
	go build -o ette

run:
	go build -o ette
	./ette
