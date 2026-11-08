# Task 4: Distributed mutual exclusion

Run three independent terminals.

```sh
go run ./node/
```

The code is prepared for three nodes. You can change this setting in the `main.go` by changing the constant `const PEERCOUNT = 3`. Each instance will wait for the user to enter the id of the process. Enter the numbers from 1 up to the peercount. The process will then wait the amout of seconds of its id untill it sends the request.
