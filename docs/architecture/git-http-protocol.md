# Git HTTP Smart Protocol Specification

## Overview

The Git HTTP smart protocol is a stateless, request/response-based protocol that enables Git operations (clone, fetch, push) over HTTP/HTTPS. It's called "smart" because the server is Git-aware and can negotiate what objects need to be transferred, unlike the older "dumb" HTTP protocol which simply served files.

## Key Characteristics

- **Stateless**: Each request is independent
- **Bidirectional**: Supports both fetch (download) and push (upload) operations
- **Efficient**: Uses packfiles and delta compression
- **Git-aware**: Server understands Git objects and can negotiate transfers
- **Browser-compatible**: Works over standard HTTP/HTTPS (but may face CORS restrictions)

## Protocol Flow

### 1. Discovery Phase

The client initiates contact by requesting the repository's capabilities and references.

#### Request Format

```
GET /repo.git/info/refs?service=git-upload-pack HTTP/1.1
Host: git.example.com
User-Agent: git/2.x.x
Git-Protocol: version=2
```

**Query Parameters:**
- `service=git-upload-pack` - For fetch/clone operations
- `service=git-receive-pack` - For push operations

#### Response Format

```
HTTP/1.1 200 OK
Content-Type: application/x-git-upload-pack-advertisement
Cache-Control: no-cache

001e# service=git-upload-pack
0000
[pkt-line formatted reference advertisement]
```

**Key Elements:**
- **Content-Type**: Must match the service type
- **Pkt-line format**: Each line is prefixed with a 4-character hex length
- **Service line**: First line announces the service
- **Flush packet**: `0000` separates sections
- **Reference advertisement**: List of refs with their SHA-1 hashes

### 2. Reference Advertisement

The server advertises all available references (branches, tags) and their current commit hashes.

#### Format

```
001e# service=git-upload-pack
0000
[capabilities line]
[zero or more ref lines]
0000
```

#### Capabilities Line

```
003fHEAD\x00multi_ack thin-pack side-band side-band-64k ofs-delta shallow
```

**Common Capabilities:**
- `multi_ack` / `multi_ack_detailed` - Acknowledges multiple "have" lines
- `thin-pack` - Server can send thin packs (with missing bases)
- `side-band` / `side-band-64k` - Multiplexed output streams
- `ofs-delta` - Use offset deltas in packfiles
- `shallow` - Support for shallow clones
- `no-progress` - Client doesn't want progress reporting
- `include-tag` - Include annotated tags
- `report-status` - Detailed push status reporting
- `delete-refs` - Allow reference deletion
- `agent=git/2.x.x` - Client/server identification

#### Reference Lines

```
003fef58a912bc... refs/heads/main
003a8f3e2a1bc... refs/heads/feature
003b1a2b3c4d5... refs/tags/v1.0.0
```

**Format:** `<length><hash> <ref-name>\n`

### 3. Service Request Phase

After discovery, the client makes a service-specific request.

## Upload-Pack Service (Fetch/Clone)

Used when the client wants to download objects from the server.

### Request

```
POST /repo.git/git-upload-pack HTTP/1.1
Host: git.example.com
Content-Type: application/x-git-upload-pack-request
Accept: application/x-git-upload-pack-result
Git-Protocol: version=2

[pkt-line formatted request body]
```

### Request Body Format

```
0032want <commit-hash>
0032want <commit-hash>
00000009done\n
```

**Want/Have Negotiation:**
1. **want**: Client lists commits it wants
2. **have**: Client lists commits it already has
3. **done**: Client signals end of negotiation
4. **NAK/ACK**: Server acknowledges or rejects

#### Negotiation Example

```
# Client sends wants
0032want abc123... multi_ack side-band-64k
0032want def456...
0000

# Client sends haves
0032have 789ghi...
0032have 012jkl...
0000

# Server responds with ACKs
0008NAK\n
# or
0030ACK 789ghi... continue
0008NAK\n

# Client signals done
0009done\n
```

### Response

The server responds with a packfile containing the requested objects.

```
HTTP/1.1 200 OK
Content-Type: application/x-git-upload-pack-result

[pkt-line formatted packfile data]
```

## Receive-Pack Service (Push)

Used when the client wants to upload objects to the server.

### Request

```
POST /repo.git/git-receive-pack HTTP/1.1
Host: git.example.com
Content-Type: application/x-git-receive-pack-request
Accept: application/x-git-receive-pack-result

[pkt-line formatted updates]
[packfile data]
```

### Request Body Format

```
# Reference updates
005b<old-hash> <new-hash> refs/heads/main\0 report-status
0000

# Packfile with new objects
PACK[packfile data]
```

**Reference Update Format:**
- `<old-hash> <new-hash> <ref-name>` - Update ref
- `0{40} <new-hash> <ref-name>` - Create ref
- `<old-hash> 0{40} <ref-name>` - Delete ref

### Response

```
HTTP/1.1 200 OK
Content-Type: application/x-git-receive-pack-result

[pkt-line formatted status report]
```

**Status Report Format:**
```
0025unpack ok
0019ok refs/heads/main
0000
```

## Pkt-Line Format

The Git protocol uses a custom framing format called "pkt-line".

### Structure

```
<4-byte hex length><payload>\n
```

- **Length**: 4 hex characters (0001-FFFF) representing total length including the 4 length bytes
- **Payload**: Actual data
- **Special values:**
  - `0000` - Flush packet (separator/terminator)
  - `0001` - Delimiter packet (protocol v2)
  - `0002` - Message packet (protocol v2)

### Examples

```
0006a\n          # Length=6, payload="a\n"
0005a            # Length=5, payload="a" (no newline)
0009done\n       # Length=9, payload="done\n"
0000             # Flush packet
```

### Encoding/Decoding

```javascript
// Encode
function encodePktLine(data) {
  if (data === null) return '0000';  // Flush packet
  const length = data.length + 4;
  return length.toString(16).padStart(4, '0') + data;
}

// Decode
function decodePktLine(line) {
  const length = parseInt(line.substring(0, 4), 16);
  if (length === 0) return null;  // Flush packet
  return line.substring(4, length);
}
```

## Packfile Format

Packfiles are the compact binary format Git uses to transfer objects.

### Structure

```
[Header: "PACK"][Version: 4 bytes][Object count: 4 bytes]
[Object 1]
[Object 2]
...
[Object N]
[Checksum: 20 bytes SHA-1]
```

### Object Entry Format

Each object in a packfile has:

```
[Type and size header (variable length)]
[Compressed data (zlib)]
```

#### Type and Size Header

Uses a variable-length encoding:
- **Byte 1 bits:**
  - Bit 7: More bytes follow (MSB)
  - Bits 6-4: Object type (3 bits)
  - Bits 3-0: Size bits (4 bits)
- **Subsequent bytes:** Size continuation (7 bits per byte, MSB indicates more bytes)

**Object Types:**
- 1: `OBJ_COMMIT`
- 2: `OBJ_TREE`
- 3: `OBJ_BLOB`
- 4: `OBJ_TAG`
- 6: `OBJ_OFS_DELTA` - Delta with offset to base
- 7: `OBJ_REF_DELTA` - Delta with SHA-1 reference to base

### Delta Objects

Git uses delta compression to save space.

#### OFS_DELTA (Offset Delta)

```
[Type and size header: type=6]
[Negative offset to base object (variable length)]
[Delta data (compressed)]
```

The offset is encoded in a special variable-length format:
```
c = byte & 0x7f;
while (byte & 0x80) {
  c = ((c + 1) << 7) | (next_byte & 0x7f);
}
offset = c;
```

#### REF_DELTA (Reference Delta)

```
[Type and size header: type=7]
[Base object SHA-1: 20 bytes]
[Delta data (compressed)]
```

### Delta Data Format

After decompression, delta data contains instructions:

```
[Source size (variable length)]
[Target size (variable length)]
[Delta instructions...]
```

**Instructions:**
- **Copy from source:** `cmd & 0x80` is set
  - Encodes offset and size to copy from base object
- **Insert new data:** `cmd & 0x80` is clear
  - cmd = number of bytes to insert
  - Followed by that many literal bytes

## Side-Band Protocol

When `side-band` or `side-band-64k` capability is active, the server multiplexes different streams.

### Format

```
<pkt-line><channel><data>
```

**Channels:**
- **1**: Packfile data
- **2**: Progress messages (stderr)
- **3**: Error messages (fatal errors)

### Example

```
001e\1[packfile bytes]
0019\2Counting objects...
0000
```

## Authentication

### HTTP Basic Auth

```
Authorization: Basic base64(username:password)
```

### Token-Based Auth

```
Authorization: Bearer <token>
```

For GitHub:
```
Authorization: token <github-personal-access-token>
```

## CORS Considerations for Browser Implementation

### Challenges

1. **Preflight requests**: OPTIONS requests for CORS
2. **Custom headers**: `Git-Protocol`, `Content-Type`
3. **Credentials**: Authentication in cross-origin requests
4. **Binary data**: Packfiles are binary

### Required CORS Headers (Server-side)

```
Access-Control-Allow-Origin: https://your-app.com
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, Git-Protocol
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

### Client-side Considerations

```javascript
fetch(url, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-git-upload-pack-request',
    'Git-Protocol': 'version=2'
  },
  credentials: 'include',  // For authentication
  mode: 'cors'
});
```

### CORS Proxy Pattern

For servers without CORS support, use a proxy:

```
Client → CORS Proxy → Git Server
```

The proxy adds necessary CORS headers.

## Error Handling

### HTTP Status Codes

- **200 OK**: Success
- **304 Not Modified**: No changes (with If-None-Match)
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Access denied
- **404 Not Found**: Repository not found
- **500 Internal Server Error**: Server error

### Git Protocol Errors

Errors in pkt-line format:
```
0028ERR error message here
```

### Common Errors

1. **CORS Error**: "No 'Access-Control-Allow-Origin' header"
   - Solution: Use CORS proxy or configure server

2. **Authentication Failed**:
   - Check credentials
   - Verify token permissions

3. **Invalid Packfile**:
   - Checksum mismatch
   - Corrupted data
   - Solution: Retry or re-clone

## Protocol Version 2

Git Protocol v2 (introduced in Git 2.18) is a more flexible protocol.

### Differences from v1

- **Capability-first**: Client requests specific capabilities
- **Command-based**: Explicit commands instead of implicit flow
- **More efficient**: Less data transferred
- **Extensible**: Easier to add new features

### Request

```
GET /repo.git/info/refs?service=git-upload-pack HTTP/1.1
Git-Protocol: version=2
```

### Commands

- `ls-refs`: List references (replaces advertisement)
- `fetch`: Fetch objects
- `push`: Push objects

## Implementation Checklist

### Phase 1: Discovery
- [ ] Implement GET info/refs request
- [ ] Parse pkt-line format
- [ ] Parse reference advertisement
- [ ] Parse capabilities

### Phase 2: Fetch/Clone
- [ ] Implement POST git-upload-pack
- [ ] Generate want/have lists
- [ ] Implement negotiation loop
- [ ] Receive and parse packfile
- [ ] Handle side-band output

### Phase 3: Packfile Processing
- [ ] Parse packfile header
- [ ] Decode object entries
- [ ] Handle delta objects (OFS_DELTA, REF_DELTA)
- [ ] Decompress objects (zlib)
- [ ] Verify checksum
- [ ] Store objects in object database

### Phase 4: Push
- [ ] Implement POST git-receive-pack
- [ ] Generate reference updates
- [ ] Create packfile with new objects
- [ ] Parse status report

### Phase 5: Error Handling
- [ ] Detect CORS issues
- [ ] Handle authentication errors
- [ ] Validate packfiles
- [ ] Implement retry logic

## References

- Git HTTP Protocol Documentation: https://git-scm.com/docs/http-protocol
- Git Transfer Protocols: https://git-scm.com/book/en/v2/Git-Internals-Transfer-Protocols
- Protocol v2: https://git-scm.com/docs/protocol-v2
- Pack Format: https://git-scm.com/docs/pack-format
- Git Source Code: https://github.com/git/git (see Documentation/technical/)

## Testing Strategy

### Unit Tests
- Pkt-line encoding/decoding
- Reference parsing
- Packfile header parsing
- Delta instruction parsing

### Integration Tests
- Fetch from public repository (e.g., small GitHub repo)
- Parse real packfiles
- Handle various ref formats
- Test error conditions

### Test Repositories
- Use small, public repositories for testing
- Create local test server with known state
- Test with different Git versions

## Performance Considerations

1. **Streaming**: Process packfiles incrementally, don't load entirely into memory
2. **Web Workers**: Parse packfiles in background thread
3. **IndexedDB**: Stream objects directly to storage
4. **Chunked Processing**: Process large packfiles in chunks
5. **Caching**: Cache parsed references and capabilities

## Security Considerations

1. **Validate URLs**: Prevent SSRF attacks
2. **Sanitize inputs**: Validate all protocol inputs
3. **Checksum verification**: Always verify packfile checksums
4. **Size limits**: Enforce maximum packfile size
5. **Rate limiting**: Implement retry backoff
6. **Credential security**: Never log credentials
7. **HTTPS only**: Prefer HTTPS for authentication

## Browser Compatibility

### Supported APIs
- Fetch API (all modern browsers)
- ArrayBuffer / Uint8Array
- TextDecoder / TextEncoder
- Streams API (for large packfiles)

### Limitations
- CORS restrictions (may need proxy)
- No SSH support (HTTP/HTTPS only)
- Memory limits for large repositories
- No git:// protocol support
