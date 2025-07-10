package t128

import (
	"bytes"
	"crypto/rc4"
	"crypto/sha1"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kdsmith18542/gordp/core"
	"github.com/kdsmith18542/gordp/glog"
	"github.com/kdsmith18542/gordp/proto/fastpath"
	"github.com/kdsmith18542/gordp/proto/mcs"
	"github.com/kdsmith18542/gordp/proto/x224"
)

// FastPathEncryptionManager handles RDP encryption for FastPath data
type FastPathEncryptionManager struct {
	encryptCipher *rc4.Cipher
	decryptCipher *rc4.Cipher
	encryptKey    []byte
	decryptKey    []byte
	seqNum        uint32
	mutex         sync.Mutex
	stats         *EncryptionStats
}

// EncryptionStats tracks encryption performance
type EncryptionStats struct {
	TotalEncrypted  int64
	TotalDecrypted  int64
	EncryptionTime  int64 // nanoseconds
	DecryptionTime  int64 // nanoseconds
	Errors          int64
	SessionKeyCount int64
}

// NewFastPathEncryptionManager creates a new FastPath encryption manager
func NewFastPathEncryptionManager() *FastPathEncryptionManager {
	return &FastPathEncryptionManager{
		stats: &EncryptionStats{},
	}
}

// SetSessionKeys sets the encryption and decryption session keys
func (em *FastPathEncryptionManager) SetSessionKeys(encryptKey, decryptKey []byte) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.encryptKey = make([]byte, len(encryptKey))
	copy(em.encryptKey, encryptKey)
	em.decryptKey = make([]byte, len(decryptKey))
	copy(em.decryptKey, decryptKey)

	// Create RC4 ciphers
	var err error
	em.encryptCipher, err = rc4.NewCipher(em.encryptKey)
	if err != nil {
		em.stats.Errors++
		return fmt.Errorf("failed to create RC4 encryption cipher: %v", err)
	}

	em.decryptCipher, err = rc4.NewCipher(em.decryptKey)
	if err != nil {
		em.stats.Errors++
		return fmt.Errorf("failed to create RC4 decryption cipher: %v", err)
	}

	em.stats.SessionKeyCount++
	glog.Debugf("FastPath encryption manager initialized with session keys")
	return nil
}

// Encrypt encrypts FastPath data using RC4
func (em *FastPathEncryptionManager) Encrypt(data []byte) ([]byte, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.encryptCipher == nil {
		return nil, fmt.Errorf("encryption cipher not initialized")
	}

	startTime := time.Now()
	defer func() {
		em.stats.EncryptionTime = time.Since(startTime).Nanoseconds()
	}()

	// Create a copy of the data to encrypt
	encrypted := make([]byte, len(data))
	copy(encrypted, data)

	// Encrypt the data in-place
	em.encryptCipher.XORKeyStream(encrypted, encrypted)

	em.stats.TotalEncrypted += int64(len(data))
	glog.Debugf("FastPath encryption: %d bytes encrypted", len(data))

	return encrypted, nil
}

// Decrypt decrypts FastPath data using RC4
func (em *FastPathEncryptionManager) Decrypt(data []byte) ([]byte, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.decryptCipher == nil {
		return nil, fmt.Errorf("decryption cipher not initialized")
	}

	startTime := time.Now()
	defer func() {
		em.stats.DecryptionTime = time.Since(startTime).Nanoseconds()
	}()

	// Create a copy of the data to decrypt
	decrypted := make([]byte, len(data))
	copy(decrypted, data)

	// Decrypt the data in-place
	em.decryptCipher.XORKeyStream(decrypted, decrypted)

	em.stats.TotalDecrypted += int64(len(data))
	glog.Debugf("FastPath decryption: %d bytes decrypted", len(data))

	return decrypted, nil
}

// GenerateSessionKeys generates session keys from the master key
func (em *FastPathEncryptionManager) GenerateSessionKeys(masterKey, clientRandom, serverRandom []byte) error {
	// Generate client-to-server key
	clientToServerKey := em.generateKey(masterKey, clientRandom, serverRandom, []byte("client-to-server"))

	// Generate server-to-client key
	serverToClientKey := em.generateKey(masterKey, clientRandom, serverRandom, []byte("server-to-client"))

	return em.SetSessionKeys(clientToServerKey, serverToClientKey)
}

// generateKey generates a session key using the RDP key derivation function
func (em *FastPathEncryptionManager) generateKey(masterKey, clientRandom, serverRandom, magic []byte) []byte {
	// RDP key derivation: SHA1(masterKey + magic + clientRandom + serverRandom)
	h := sha1.New()
	h.Write(masterKey)
	h.Write(magic)
	h.Write(clientRandom)
	h.Write(serverRandom)
	return h.Sum(nil)
}

// GetStats returns encryption statistics
func (em *FastPathEncryptionManager) GetStats() *EncryptionStats {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	stats := *em.stats // Copy to avoid race conditions
	return &stats
}

// ResetStats resets encryption statistics
func (em *FastPathEncryptionManager) ResetStats() {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.stats = &EncryptionStats{}
}

// IsInitialized returns true if the encryption manager is properly initialized
func (em *FastPathEncryptionManager) IsInitialized() bool {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	return em.encryptCipher != nil && em.decryptCipher != nil
}

type PDU interface {
	iPDU()
	Read(r io.Reader) PDU
	Serialize() []byte
	Type() uint16
}

type DataPDU interface {
	iDataPDU()
	Read(r io.Reader) DataPDU
	Serialize() []byte
	Type2() uint8
}

var pduMap = map[uint16]PDU{
	PDUTYPE_DEMANDACTIVEPDU:  &TsDemandActivePduData{},
	PDUTYPE_CONFIRMACTIVEPDU: &TsConfirmActivePduData{},
	PDUTYPE_DEACTIVATEALLPDU: nil,
	PDUTYPE_DATAPDU:          &TsDataPduData{},
	PDUTYPE_SERVER_REDIR_PKT: nil,
}

var pduMap2 = map[uint8]DataPDU{
	PDUTYPE2_SYNCHRONIZE:                 &TsSynchronizePduData{},
	PDUTYPE2_CONTROL:                     &TsControlPDU{},
	PDUTYPE2_FONTMAP:                     &TsFontMapPDU{},
	PDUTYPE2_SET_ERROR_INFO_PDU:          &TsSetErrorInfoPDU{},
	PDUTYPE2_SAVE_SESSION_INFO:           &TsSaveSessionInfoPDU{},
	PDUTYPE2_BITMAPCACHE_PERSISTENT_LIST: &TsBitmapCachePersistentListPDU{},
	PDUTYPE2_BITMAPCACHE_ERROR_PDU:       &TsBitmapCacheErrorPDU{},
}

func readPDU(r io.Reader, typ uint16) PDU {
	if _, ok := pduMap[typ]; !ok {
		core.Throw(fmt.Errorf("invalid pdu type: %v", typ))
	}
	return pduMap[typ].Read(r)
}

func readMcsSdin(r io.Reader) []byte {
	var mcsSDin mcs.ReceiveDataResponse
	channelId, data := mcsSDin.Read(r)
	glog.Debugf("read pdu from channel: %v, %x", channelId, data)
	return data
}

func ReadExpectedPDU(r io.Reader, typ uint16) PDU {
	r = bytes.NewReader(readMcsSdin(r))
	header := TsShareControlHeader{}
	header.Read(r)
	glog.Debugf("share ctrl header: %+v", header)
	core.ThrowIf(header.PDUType != typ, "not expected PDU type")
	return readPDU(r, typ)
}

func ReadPDU(r io.Reader) PDU {
	r = bytes.NewReader(readMcsSdin(r))
	header := TsShareControlHeader{}
	header.Read(r)
	return readPDU(r, header.PDUType)
}

func WritePDU(w io.Writer, userId uint16, pdu PDU) {
	data := pdu.Serialize()
	header := TsShareControlHeader{
		PDUType:     pdu.Type(),
		PDUSource:   userId,
		TotalLength: uint16(len(data) + 6),
	}
	glog.Debugf("pdu.Serialize: %v - %x", len(data), data)

	mcsSDrq := mcs.NewSendDataRequest(userId, mcs.MCS_CHANNEL_GLOBAL)
	data = mcsSDrq.Serialize(append(header.Serialize(), data...))
	x224.Write(w, data)
}

func ReadExpectedDataPDU(r io.Reader, typ2 uint8) DataPDU {
	pdu := ReadExpectedPDU(r, PDUTYPE_DATAPDU).(*TsDataPduData)
	core.ThrowIf(pdu.Header.PDUType2 != typ2, "invalid pdu type2")
	return pdu.Pdu
}

func WriteDataPdu(w io.Writer, userId uint16, shareId uint32, pdu DataPDU) {
	WritePDU(w, userId, NewDataPdu(pdu, shareId))
}

// Global FastPath encryption manager
var fastPathEncryptionManager = NewFastPathEncryptionManager()

// SetFastPathEncryptionManager sets the global FastPath encryption manager
func SetFastPathEncryptionManager(manager *FastPathEncryptionManager) {
	fastPathEncryptionManager = manager
}

// GetFastPathEncryptionManager returns the global FastPath encryption manager
func GetFastPathEncryptionManager() *FastPathEncryptionManager {
	return fastPathEncryptionManager
}

func ReadFastPathPDU(r io.Reader) PDU {
	fp := fastpath.Read(r)

	// Handle encryption if present
	if fp.Header.EncryptionFlags != 0 {
		glog.Debugf("FastPath encryption detected (flags: %d), decrypting data", fp.Header.EncryptionFlags)

		if !fastPathEncryptionManager.IsInitialized() {
			core.Throw(fmt.Errorf("FastPath encryption required but encryption manager not initialized"))
		}

		// Decrypt the data
		decryptedData, err := fastPathEncryptionManager.Decrypt(fp.Data)
		if err != nil {
			core.Throw(fmt.Errorf("failed to decrypt FastPath data: %v", err))
		}

		fp.Data = decryptedData
		glog.Debugf("FastPath data decrypted: %d bytes", len(decryptedData))
	}

	glog.Debugf("analyse FastPathPDU")
	return (&TsFpUpdatePDU{}).Read(bytes.NewReader(fp.Data))
}

func WriteFastPathInputPDU(w io.Writer, pdu *TsFpInputPdu) {
	data := pdu.Serialize()

	// Check if encryption is enabled
	if fastPathEncryptionManager.IsInitialized() {
		glog.Debugf("Encrypting FastPath input data: %d bytes", len(data))

		// Encrypt the data
		encryptedData, err := fastPathEncryptionManager.Encrypt(data)
		if err != nil {
			core.Throw(fmt.Errorf("failed to encrypt FastPath data: %v", err))
		}

		data = encryptedData

		// Set encryption flags in the header
		header := fastpath.Header{
			EncryptionFlags: 1, // Basic encryption
			Length:          len(data),
		}
		header.Write(w)
		core.WriteFull(w, data)
	} else {
		// No encryption
		fastpath.Write(w, data)
	}
}

// WriteFastPathPDU writes a FastPath PDU with optional encryption
func WriteFastPathPDU(w io.Writer, pdu PDU, encrypt bool) {
	data := pdu.Serialize()

	if encrypt && fastPathEncryptionManager.IsInitialized() {
		glog.Debugf("Encrypting FastPath PDU: %d bytes", len(data))

		// Encrypt the data
		encryptedData, err := fastPathEncryptionManager.Encrypt(data)
		if err != nil {
			core.Throw(fmt.Errorf("failed to encrypt FastPath PDU: %v", err))
		}

		data = encryptedData

		// Set encryption flags in the header
		header := fastpath.Header{
			EncryptionFlags: 1, // Basic encryption
			Length:          len(data),
		}
		header.Write(w)
		core.WriteFull(w, data)
	} else {
		// No encryption
		fastpath.Write(w, data)
	}
}
