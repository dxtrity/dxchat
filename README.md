# dxchat
**dxchat** is a **TCP** based terminal chat application consisting of a server and a client.
The server and client are independant applications that can send and receive messages to eachother.
The chat application supports multiple clients, backing up messages and more.

## Server Features

**Multiple Clients**<br>
By default the server is capable of hosting up to 8 clients.
This will be able to be configured in a config file for the server.

**Message Backup**<br>
The server backs up any messages sent to it in 10 minute intervals.

## Client Features

**Terminal UI**<br>
The client features a UI in the terminal consisting of a input field and a content field for messages. It also support colours which are managed by the client and sent by the server.

**Commands**<br>
The client features commands such as `:quit` to quit the application.

## Todo list

- [ ] Add `:ctx`: Opens **Context Manager**
- [ ] Add `:rooms`: Opens available **rooms**
- [ ] Add config files for client and server
- [ ] Add theme support