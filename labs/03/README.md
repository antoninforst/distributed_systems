# 1. What are packages in your implementation? What data structure do you use to transmit data and meta-data?

Packages are objects of type Message, which is two bolean flags (ACK, SYN) and two ints with the sequence numbers and string representing the data.

# 2. Does your implementation use threads or processes? Why is it not realistic to use threads?

It uses threads. Threads are meant to divide work inside one computer. TCP is developed to happen over network among different computers.

# 3. In case the network changes the order in which messages are delivered, how would you handle message re-ordering?

I would reorder them based on the sqeuential numbers.

# 4. In case messages can be delayed or lost, how does your implementation handle message loss?

Optional extention. It is not implemented. It never goes into the error section.

# 5. Why is the 3-way handshake important?

The communication can be only onesided, so the server and the client have to assure, that they both listen and hear each other. Without the last step the server would not know, that the clinet hear it.
