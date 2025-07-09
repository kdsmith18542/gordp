# Network Level Authentication (NLA) Implementation

This package implements Network Level Authentication (NLA) for the RDP protocol, providing secure authentication using NTLMv2 and CredSSP.

## Features

### Core NLA Implementation
- **NTLMv2 Authentication**: Full implementation of the NTLMv2 handshake
- **CredSSP Support**: Credential Security Support Provider for secure credential exchange
- **Channel Binding Token (CBT)**: Protection against man-in-the-middle attacks

### Channel Binding Token (CBT)

Channel Binding Token is a security feature that binds the NLA authentication to the underlying TLS channel, preventing man-in-the-middle attacks.

#### How it works:
1. **Certificate Hash**: The client computes a SHA256 hash of the server's TLS certificate
2. **NTLM Integration**: The hash is included in the NTLMv2 client challenge as an AVPair
3. **Server Verification**: The server verifies the hash matches its actual certificate

#### Implementation Details:
- **RFC 5929 Compliance**: Implements the "tls-server-end-point" channel binding type
- **Automatic Detection**: Automatically enabled for SSL/HYBRID protocol connections
- **Fallback Support**: Gracefully handles cases where channel binding is not available

#### Usage:
```go
// Channel binding is automatically enabled for SSL/HYBRID connections
client := gordp.NewClient(&gordp.Option{
    Addr:     "server:3389",
    UserName: "user",
    Password: "password",
})

// The NLA handshake will automatically include channel binding
err := client.Connect()
```

### Message Types

#### NegotiateMessage
- Initiates the NTLM handshake
- Specifies supported features and capabilities
- Sent by the client to start authentication

#### ChallengeMessage
- Server response to negotiation
- Contains server challenge and target information
- Includes timestamp and other security parameters

#### AuthenticateMessage
- Client response to challenge
- Contains computed responses and session keys
- Includes channel binding token when available

### Security Features

#### Session Key Derivation
- **Client/Server Signing Keys**: For message integrity
- **Client/Server Sealing Keys**: For message confidentiality
- **RC4 Encryption**: For secure communication

#### Message Integrity Check (MIC)
- HMAC-MD5 based integrity protection
- Covers all NTLM messages in the handshake
- Prevents tampering with authentication data

### AVPairs (Attribute-Value Pairs)

The implementation supports various AVPairs for extended functionality:

- `MsvAvTimestamp`: Server timestamp for replay protection
- `MsvAvTargetName`: Target server information
- `MsvChannelBindings`: Channel binding token (CBT)
- `MsvAvEOL`: End-of-list marker

### Error Handling

The implementation includes comprehensive error handling:
- **Protocol Validation**: Ensures message types and formats are correct
- **Connection State**: Validates TLS connection for channel binding
- **Graceful Degradation**: Falls back gracefully when features are unavailable

## Testing

The package includes comprehensive tests:
- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end authentication flow
- **Channel Binding Tests**: Specific CBT functionality validation

Run tests with:
```bash
go test -v ./proto/nla
```

## Security Considerations

1. **Channel Binding**: Always enabled for SSL/HYBRID connections to prevent MITM attacks
2. **Session Keys**: Properly derived and used for all subsequent communication
3. **Certificate Validation**: TLS certificate hash is included in authentication
4. **Replay Protection**: Timestamps and challenges prevent replay attacks

## References

- [MS-NLMP](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/99d90ff4-957f-4c8a-80e4-5bfe5a9a9832): NTLM Authentication Protocol
- [MS-CSSP](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cssp/85f57821-40bb-46aa-bfcb-ba9590b8fc30): Credential Security Support Provider
- [RFC 5929](https://tools.ietf.org/html/rfc5929): Channel Bindings for TLS 