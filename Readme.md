# Peerster

This is just a quick summary of the different homeworks.

## Homework 1
1. Creating a UDP (datagram) socket to communicate with other nodes.
2. Implementing a simple gossip algorithm to distribute messages among directly and
indirectly connected nodes.
3. Constructing a simple user interface for viewing and entering messages.

### Gossiping in Peerster

Nodes can join and leave the gossip protocol. When a node :
* joins -> contact info from  a few other nodes accessible.  
* receives a message, it learns the address from the sender.

### Peerster design
Each node acts as a gossiper. The client contacts the gossiper through an API to send, receive, list messages etc.  
The gossiper node communicates with other peers on the `gossipPort` , and with clients on the `UIPort`. Hence the gossiper listens on two different ports.

## Part 1: Network Programming in Go

Implement broadcasting via UDP. Create two programs: The first to initialize the gossiper and the second is the client API.