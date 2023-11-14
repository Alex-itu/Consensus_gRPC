# Consensus_gRPC
# How to run the program
First you should open a couple of terminal windows. Next step is to paste these into the terminal

go run .\node.go -name alice -port 5000 -portfor localhost:5010 -portback localhost:5020 -hasToken true /n

go run .\node.go -name bob -port 5010 -portfor localhost:5020 -portback localhost:5000 /n

go run .\node.go -name charlie -port 5020 -portfor localhost:5000 -portback localhost:5010 /n


To explain the arguments a bit:
-name is the name of the client node which gets printed in the log and terminal
-port is the port of the client you are currently creating in the terminal window
-portfor is the forward client port in the token ring aka the guy you pass the token to after you are done
-portback is the backward clients port which is the client you recieve tokens from 
-hasToken is a boolean modifier which dictates who starts with the token. !ONLY 1 CLIENT SHOULD EVER HAVE A TOKEN!
