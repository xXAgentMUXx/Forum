FROM golang:1.24-bullseye

# Install dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    wget \
    libssl-dev \
    tcl \
    autoconf \
    automake \
    libtool \
    pkg-config \
    sqlite3

# Build SQLCipher from source
WORKDIR /sqlcipher
RUN wget https://github.com/sqlcipher/sqlcipher/archive/refs/tags/v4.5.6.tar.gz && \
    tar -xzf v4.5.6.tar.gz && cd sqlcipher-4.5.6 && \
    ./configure --enable-tempstore=yes CFLAGS="-DSQLITE_HAS_CODEC" --with-crypto-lib=openssl && \
    make && make install

# Set up environment for Go to link with SQLCipher
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-I/usr/local/include"
ENV CGO_LDFLAGS="-L/usr/local/lib -lsqlcipher"

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -tags "sqlite_see" -o main ./server/main.go

CMD ["./main"]
