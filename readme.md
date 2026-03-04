# Для генерации protobuf используем 
https://buf.build/product/cli

## команда установки от автора видео 
```
go get -tool github.com/bufbuild/buf/cmd/buf@v1.50.0
```

## команда с последней версией buf
```
go install github.com/bufbuild/buf/cmd/buf@latest 
```

## ссылку на гитхаб последняя версия 
https://github.com/bufbuild/buf/releases


Далее используем команду для инициализации protobuf
команда сгенерирует файл buf.yaml
```
buf config init
```

после создать файл buf.gen.yaml будут храниться настройки связаные с генерацией кода 

после обновим
появится файл buf.lock - он сгенерировал файл журнала с зависимостями 
```
buf dep update
```
Запуск команды для генерации файлов proto
```
buf generate --template buf.gen.yaml
```
или команда 
```
make generate-proto
```

команда для линтинга протофайлов 
```
buf lint --config buf.yaml
```
или 
```
make lint-proto
```