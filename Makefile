
lint-proto::
	buf lint --config buf.yaml

generate-proto::
	buf generate --template buf.gen.yaml
