# DistributionServer
#### DistributionServer is used to route messages and other information between [EndServer](https://github.com/Adarsh-Kmt/EndServer) containers.
-----
## Features
- Stores user to end server mapping in a redis database.
- Uses gRPC to route messages between users connected to different end server containers.
- Implemented mutual TLS between distribution server container and end server container, which enables authentication of both parties, and encryption of messages exchanged.
- Leveraged multistage docker builds to reduce the size of the image by 94.78% (504.17 MB -> 26.3 MB)
