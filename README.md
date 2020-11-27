# 檢查專案是否有中文
- 把專案clone下來
- go build -o fops main.go

#### 檢查folder裡有中文的檔案 

```shell script
fops check_chinese --folder <folder> --ignore_folder <不需要檢查的folder>
```

#### 找出 const & var 裡有中文的參數

```shell script
fops get_parameter --folder <folder> --ignore_folder <不需要檢查的folder>
```

#### 找出檔案裡 import 的 package

```shell script
fops get_import_package --folder <folder> --ignore_folder <不需要檢查的folder,多個可用逗號隔開>
```

#### 範例

```shell script
fops check_chinese --folder athena --ignore_folder build,doc
```