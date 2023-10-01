## How to import third party package?
1. Just import the package in your code
2. run `$ go mod init <filename>`, it will then create a go.mod file
3. run `$ go mod tidy`, this command checks the imports used in your program and fetches the module if not fetched already.