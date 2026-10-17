.PHONY: run api web test build

api:
	go run .

web:
	cd frontend && npm run dev

test:
	go test ./services/...

build:
	cd frontend && npm run build
	go build -o bin/api .

sample-import:
	curl -s -F file=@sample_data/tinkoff_sample.csv \
	  -F account_id=1 \
	  -F transfer_account_id=2 \
	  -F bank=tinkoff \
	  http://localhost:8443/api/v1/import | jq .
