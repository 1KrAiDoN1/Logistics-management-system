proto:
	protoc -I api/protobuf/ \
		api/protobuf/auth_service/auth_service.proto \
		--go_out=./api/protobuf \
		--go_opt=paths=source_relative \
		--go-grpc_out=./api/protobuf \
		--go-grpc_opt=paths=source_relative

# 2. папки, где лежат прото файлы
# 3. полный путь к прото файлу, который нужно сгенерировать
# 4. папка, куда будут сгенерированы go файлы
# 5. параметры для генерации go файлов
# 6. папка, куда будут сгенерированы go файлы
# 7. параметры для генерации grpc файлов