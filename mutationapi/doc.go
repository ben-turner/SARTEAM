/*
Package mutationapi provides a system for managing mutations to a shared state,
and communicating them with sources such as clients and files.

In general, mutations come from a *Conn, which may be a websocket connection, a
file, or some other source. Once the *Conn is created, mutations can be consumed
from it either by calling the Receive method, or by using the Pipe function to
send them to a channel. Mutations can then be applied to a state, forwarded to
other *Conns (perhaps in a ConnSet) or reverted if necessary.

The expected usage is to have a ConnSet which holds all the active connections,
and a channel which receives mutations from all of them. The mutations can then
be applied to the state, and the ConnSet can be used to broadcast the mutation
to synchronize the state with all the other *Conns. If the mutation cannot be
applied, the inverse of the mutation can be sent to the original *Conn to revert
it. In this case, an error should also be returned to the client explaining why
the mutation failed.
*/
package mutationapi
