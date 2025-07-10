# RDP Protocol Documentation

## Overview

This document provides detailed information about the Remote Desktop Protocol (RDP) implementation in GoRDP. It covers the protocol layers, message formats, and implementation details.

## Table of Contents

1. [Protocol Architecture](#protocol-architecture)
2. [Connection Flow](#connection-flow)
3. [Protocol Layers](#protocol-layers)
4. [Message Formats](#message-formats)
5. [Security](#security)
6. [Virtual Channels](#virtual-channels)
7. [Input Handling](#input-handling)
8. [Graphics](#graphics)
9. [Audio](#audio)
10. [Device Redirection](#device-redirection)

## Protocol Architecture

The RDP protocol is implemented as a layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│                    Virtual Channels                         │
├─────────────────────────────────────────────────────────────┤
│                    Multipoint Communication Service (MCS)   │
├─────────────────────────────────────────────────────────────┤
│                    Security Layer                           │
├─────────────────────────────────────────────────────────────┤
│                    X.224 Connection Layer                   │
├─────────────────────────────────────────────────────────────┤
│                    TPKT Transport Layer                     │
└─────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

- **TPKT**: Transport layer providing packet framing
- **X.224**: Connection-oriented transport protocol
- **Security**: Encryption, authentication, and integrity
- **MCS**: Multipoint communication and channel management
- **Virtual Channels**: Application-specific data channels
- **Application**: RDP-specific protocol data units (PDUs)

## Connection Flow

### 1. Connection Establishment

```
Client                    Server
  |                        |
  |-- Connection Request ->|
  |                        |
  |<- Connection Confirm --|
  |                        |
```

### 2. Basic Settings Exchange

```
Client                    Server
  |                        |
  |-- MCS Connect Initial ->|
  |                        |
  |<- MCS Connect Response-|
  |                        |
```

### 3. Channel Connection

```
Client                    Server
  |                        |
  |-- Erect Domain Request->|
  |                        |
  |-- Attach User Request ->|
  |                        |
  |<- Attach User Confirm --|
  |                        |
  |-- Channel Join Request->|
  |                        |
  |<- Channel Join Confirm-|
  |                        |
```

### 4. Security Commencement

```
Client                    Server
  |                        |
  |-- Security Exchange --->|
  |                        |
```

### 5. Secure Settings Exchange

```
Client                    Server
  |                        |
  |-- Client Info -------->|
  |                        |
```

### 6. Licensing

```
Client                    Server
  |                        |
  |-- License Request ---->|
  |                        |
  |<- License Response ----|
  |                        |
```

### 7. Capabilities Exchange

```
Client                    Server
  |                        |
  |-- Demand Active ------>|
  |                        |
  |<- Confirm Active ------|
  |                        |
```

### 8. Connection Finalization

```
Client                    Server
  |                        |
  |-- Synchronize -------->|
  |                        |
  |-- Control Request ---->|
  |                        |
  |-- Font List ---------->|
  |                        |
```

### 9. Data Exchange

```
Client                    Server
  |                        |
  |<-- Bitmap Updates ----|
  |                        |
  |-- Input Events ------>|
  |                        |
```

## Protocol Layers

### TPKT Layer

The TPKT (Transport Protocol Data Unit) layer provides packet framing for the RDP protocol.

#### TPKT Header Format

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Version    |   Reserved    |        Length                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Version**: Always 3 for RDP
- **Reserved**: Must be 0
- **Length**: Total length of the TPKT packet including header

#### Implementation

```go
type TPKTHeader struct {
    Version  uint8
    Reserved uint8
    Length   uint16
}
```

### X.224 Layer

The X.224 layer provides connection-oriented transport services.

#### X.224 Header Format

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Length     |    PDU Type   |    Dst Reference              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Src Reference              |    Class Option               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### PDU Types

- **Connection Request (0xE0)**: Initial connection request
- **Connection Confirm (0xD0)**: Connection acceptance
- **Disconnect Request (0x80)**: Connection termination
- **Data (0xF0)**: Data transfer

#### Implementation

```go
type X224Header struct {
    Length        uint8
    PDUType       uint8
    DstReference  uint16
    SrcReference  uint16
    ClassOption   uint8
}
```

### Security Layer

The security layer provides encryption, authentication, and integrity services.

#### Security Header Format

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Security Header Type       |    Security Header Flags      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Security Header Length     |    Security Header Version    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Security Types

- **RDP Security (0x0000)**: Standard RDP encryption
- **SSL Security (0x0001)**: SSL/TLS encryption
- **CredSSP Security (0x0002)**: CredSSP authentication

#### Implementation

```go
type SecurityHeader struct {
    Type    uint16
    Flags   uint16
    Length  uint16
    Version uint16
}
```

### MCS Layer

The Multipoint Communication Service (MCS) layer provides channel management and multipoint communication.

#### MCS PDU Types

- **Connect Initial**: Initial connection request
- **Connect Response**: Connection response
- **Erect Domain Request**: Domain establishment
- **Attach User Request**: User attachment
- **Attach User Confirm**: User attachment confirmation
- **Channel Join Request**: Channel joining
- **Channel Join Confirm**: Channel joining confirmation
- **Send Data Request**: Data transmission
- **Send Data Indication**: Data reception

#### Implementation

```go
type MCSPDU struct {
    Type    uint8
    Length  uint16
    Data    []byte
}
```

## Message Formats

### RDP Negotiation

#### Connection Request

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Type       |    Flags      |    Length                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Requested Protocols                                        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Connection Confirm

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Type       |    Flags      |    Length                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Selected Protocol                                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Client Info

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Length     |    Type       |    Flags                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Domain Length              |    Domain                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    User Length                |    User                       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Password Length            |    Password                   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Program Length             |    Program                    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Directory Length           |    Directory                  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Security

### Authentication Methods

1. **Standard RDP Security**: Basic encryption
2. **Enhanced RDP Security**: SSL/TLS encryption
3. **Network Level Authentication (NLA)**: CredSSP authentication

### NLA Authentication Flow

```
Client                    Server
  |                        |
  |-- Negotiate Message -->|
  |                        |
  |<- Challenge Message ---|
  |                        |
  |-- Authenticate Message>|
  |                        |
```

### Encryption

- **RDP Encryption**: RC4 with 40-bit, 56-bit, or 128-bit keys
- **SSL/TLS**: Standard SSL/TLS encryption
- **FIPS**: Federal Information Processing Standards compliance

## Virtual Channels

### Static Virtual Channels

Predefined channels for specific functionality:

- **cliprdr**: Clipboard redirection
- **rdpsnd**: Audio redirection
- **rdpdr**: Device redirection
- **drdynvc**: Dynamic virtual channels

### Dynamic Virtual Channels

User-defined channels for custom applications.

#### Channel Creation

```
Client                    Server
  |                        |
  |-- Create Request ----->|
  |                        |
  |<- Create Response -----|
  |                        |
  |-- Open Request ------->|
  |                        |
  |<- Open Response ------|
  |                        |
```

#### Data Transfer

```
Client                    Server
  |                        |
  |-- Data Message ------>|
  |                        |
  |<- Data Message -------|
  |                        |
```

## Input Handling

### Keyboard Input

#### FastPath Keyboard Event

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Event Header         |    Key Code     |    Key State      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Unicode Keyboard Event

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Event Header         |    Unicode Code |    Key State      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Mouse Input

#### FastPath Mouse Event

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Event Header         |    Pointer Flags|    X Position     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Y Position           |
+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Extended Mouse Event

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Event Header         |    Pointer Flags|    X Position     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Y Position           |    Additional Data                  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Graphics

### Bitmap Updates

#### Bitmap Data

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Destination Left     |    Destination Top                  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Destination Right    |    Destination Bottom               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Width                |    Height                           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Bits Per Pixel       |    Compression                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Bitmap Data Length   |    Bitmap Data                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Compression Types

- **RDP6**: RDP 6.0 compression
- **RLE**: Run-length encoding
- **JPEG**: JPEG compression
- **PNG**: PNG compression

### Surface Commands

#### Surface Command Header

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Command Type         |    Command Length                   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Command Types

- **Set Surface Bits**: Update surface with bitmap data
- **Frame Marker**: Mark frame boundaries
- **Stream Surface Bits**: Stream bitmap data
- **Solid Fill**: Fill area with solid color
- **Create Surface**: Create new surface
- **Delete Surface**: Delete surface

## Audio

### Audio Formats

- **PCM**: Pulse Code Modulation
- **ADPCM**: Adaptive Differential PCM
- **DVI**: Digital Video Interactive
- **GSM**: Global System for Mobile Communications

### Audio Redirection

#### Audio Format Negotiation

```
Client                    Server
  |                        |
  |-- Audio Formats ----->|
  |                        |
  |<- Audio Formats ------|
  |                        |
```

#### Audio Data Transfer

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Audio Format         |    Audio Data Length                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Audio Data                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Device Redirection

### Device Types

- **Printer**: Printer redirection
- **Drive**: File system redirection
- **Port**: Serial/parallel port redirection
- **Smart Card**: Smart card reader redirection

### Device Announcement

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Device Type          |    Device ID                        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Preferred DOS Name   |    Device Data Length               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Device Data                                              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Device I/O Request

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Device ID            |    File ID                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Completion ID        |    Major Function                   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Minor Function       |    Data Length                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Data                                                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Implementation Notes

### Error Handling

- All protocol errors are wrapped with context
- Invalid packet formats are logged and handled gracefully
- Connection timeouts are configurable
- Retry mechanisms for transient failures

### Performance Considerations

- Bitmap caching for improved performance
- Compression to reduce bandwidth usage
- Efficient memory management
- Concurrent processing where possible

### Security Considerations

- Certificate validation
- Credential protection
- Encryption of sensitive data
- Input validation and sanitization

## References

- [MS-RDPBCGR]: Remote Desktop Protocol: Basic Connectivity and Graphics Remoting
- [MS-RDPEDISP]: Remote Desktop Protocol: Display Update Virtual Channel Extension
- [MS-RDPEFS]: Remote Desktop Protocol: File System Virtual Channel Extension
- [MS-RDPEGFX]: Remote Desktop Protocol: Graphics Pipeline Extension
- [MS-RDPEI]: Remote Desktop Protocol: Input Virtual Channel Extension
- [MS-RDPEMT]: Remote Desktop Protocol: Multitransport Extension
- [MS-RDPEUDP]: Remote Desktop Protocol: UDP Transport Extension 