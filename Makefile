proto_clean:
	rm -rfv app/pb

proto_gen:
	cd app
	mkdir pb
	protoc -I proto/ --go_out=paths=source_relative:pb proto/*.proto
	cd ..

build:
	go build -o ette

run:
	go build -o ette
	./ette
