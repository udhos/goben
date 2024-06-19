all: ca.crt ca.key client.crt client.key

ca.crt ca.key: openssl.cnf
	# create a self-signed CA
	# - add your subject alternative names (server hostnames) to the config at the end
	openssl req -x509 -new -nodes -keyout ca.key -out ca.crt -days 3650 -config openssl.cnf

client.key:
	openssl genrsa -out client.key 2048

client.crt: client.key ca.crt ca.key
	openssl req -new -key client.key -out client.csr -subj "/C=AT/ST=Vienna/L=test/O=test/OU=test_ou/CN=client1"
	openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 3650 -sha256
	rm client.csr
	openssl verify -CAfile ca.crt client.crt
