# wg-drill

Inspired and reference [wgcf](https://github.com/jwhited/wgsd).

Simplify for dn11.

## Introducing

Use wireguard's peer as stun and exchange target endpoint.

## install

Download package or build yourself

Running:

````
wg-drill-[client/server] install
````

Useage:

reference config file below and [jwhited's blog](https://www.jordanwhited.com/posts/wireguard-endpoint-discovery-nat-traversal/)

````
[Interface]
PrivateKey = <Your Private Key>
ListenPort = <Listen Port>

# start with interface
PostUp = wg-drill-client up <Your Interface name>
# stop with interface
preDown = wg-drill-client down <Your Interface name>#



#Target
[Peer]
PublicKey = <Target Pubkey>

#Server
[Peer]
Endpoint = <Server Listen Port>
PublicKey = <Server Pubkey>
PersistentKeepalive = 3
````

## ToDo
- [ ] Simplfy ways to get endpoint(We use http temporary and need to config server endpoint.This version would be terrible while having multi server) 
- [ ] Better ways to close server
- [ ] Log