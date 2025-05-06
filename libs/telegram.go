package libs

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"go-telegram-binance/config"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Singleton TelegramClient
type TelegramClient struct {
	client       *telegram.Client
	dispatcher   tg.UpdateDispatcher // Bukan pointer (*tg.UpdateDispatcher)
	targetChatID int64
	accessHash   int64
}

var (
	instance *TelegramClient
	once     sync.Once
	ctx      context.Context
)

// InitTelegramClient -> Membuat singleton TelegramClient
func InitTelegramClient(apiID int, apiHash string) *TelegramClient {
	once.Do(func() {
		ctx = context.Background()
		dispatcher := tg.NewUpdateDispatcher()
		sessionFile := "session.json"
		storage := &session.FileStorage{
			Path: sessionFile,
		}

		instance = &TelegramClient{
			client:     telegram.NewClient(apiID, apiHash, telegram.Options{UpdateHandler: &dispatcher, SessionStorage: storage}),
			dispatcher: dispatcher, // Tidak perlu pakai pointer
		}

		go func() {
			// Jalankan Telegram Client lebih dulu
			err := instance.client.Run(ctx, func(ctx context.Context) error {
				log.Println("✅ Telegram Client Berjalan...")

				status, err := instance.client.Auth().Status(ctx)
				if err != nil {
					return fmt.Errorf("gagal cek status login: %w", err)
				}

				if !status.Authorized {
					log.Println("⚠️ Anda belum login. Silakan login terlebih dahulu.")
					var phone string
					fmt.Print("Masukkan nomor HP Anda (misalnya: +6281234567890): ")
					fmt.Scanln(&phone)

					var password string
					fmt.Print("Masukkan password jika memiliki password : ")
					fmt.Scanln(&password)

					codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
						// NB: Use "golang.org/x/crypto/ssh/terminal" to prompt password.
						fmt.Print("Enter code: ")
						code, err := bufio.NewReader(os.Stdin).ReadString('\n')
						if err != nil {
							return "", err
						}
						return strings.TrimSpace(code), nil
					}

					if err := auth.NewFlow(
						auth.Constant(phone, password, auth.CodeAuthenticatorFunc(codePrompt)),
						auth.SendCodeOptions{},
					).Run(ctx, instance.client.Auth()); err != nil {
						panic(err)
					}
				}

				// Resolve username to get channel information
				resolved, err := instance.client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
					Username: config.AppConfig.TeleChannelName,
				})
				if err != nil {
					log.Fatalf("Gagal mendapatkan ChannelID: %v", err)
				}

				var channel *tg.InputPeerChannel
				for _, ch := range resolved.Chats {
					if c, ok := ch.(*tg.Channel); ok {
						channel = &tg.InputPeerChannel{
							ChannelID:  c.ID,
							AccessHash: c.AccessHash,
						}
						break
					}
				}

				if channel == nil {
					log.Fatalf("failed to get channel information")
				}
				if err != nil {
					log.Fatalf("❌ Gagal mendapatkan ChannelID: %v", err)
				}
				instance.targetChatID = channel.ChannelID
				instance.accessHash = channel.AccessHash
				log.Printf("🎯 Target Channel ID: %d", channel.ChannelID)

				// Pasang handler
				instance.setupMessageHandler()

				// Tunggu sampai context selesai
				<-ctx.Done()
				return nil
			})
			if err != nil {
				log.Fatalf("❌ Gagal menjalankan Telegram Client: %v", err)
			}
		}()
	})

	return instance
}

// GetInstance -> Mengembalikan instance TelegramClient yang sudah dibuat
func GetInstance() *TelegramClient {
	if instance == nil {
		panic("TelegramClient belum diinisialisasi! Panggil InitTelegramClient terlebih dahulu.")
	}
	return instance
}

// setupMessageHandler -> Menangkap pesan dari channel tertentu
func (t *TelegramClient) setupMessageHandler() {
	t.dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		if msg, ok := update.Message.(*tg.Message); ok {
			if peer, ok := msg.PeerID.(*tg.PeerChannel); ok {
				// Filter hanya pesan dari channel tertentu
				if peer.ChannelID == t.targetChatID {
					fmt.Printf("📩 Pesan dari channel %d: %s\n", peer.ChannelID, msg.Message)
					// Parsing teks menjadi JSON
					marketData, err := ParseTextToJSON(msg.Message)
					if err != nil {
						fmt.Println("Error parsing text:", err)
					}

					if marketData.LargestVolume != nil {
						// Konversi ke JSON
						_, err := json.MarshalIndent(marketData, "", "  ")
						if err != nil {
							fmt.Println("Error converting to JSON:", err)
						}

						binance := config.GetInstance()

						for _, topGainer := range marketData.BinanceFutures.TopGainers {
							err = binance.CreateOrder(topGainer.Name, "BUY")
							if err != nil {
								replyText := fmt.Sprintf("❌ Order %s Buy Filed = %v", topGainer.Name, err)
								t.ReplyToMessage(ctx, msg.ID, replyText)

								continue
							} else {
								replyText := fmt.Sprintf("✅ Order %s Buy Succesfully", topGainer.Name)
								t.ReplyToMessage(ctx, msg.ID, replyText)

							}
						}

						for _, topLosers := range marketData.BinanceFutures.TopLosers {
							binance.CreateOrder(topLosers.Name, "SELL")
							if err != nil {
								replyText := fmt.Sprintf("❌ Order %s Sell Failed = %v", topLosers.Name, err)
								t.ReplyToMessage(ctx, msg.ID, replyText)
								continue
							} else {
								replyText := fmt.Sprintf("✅ Order %s Sell Succesfully", topLosers.Name)
								t.ReplyToMessage(ctx, msg.ID, replyText)
							}
						}

						// 🔥 Membalas pesan
						// replyText := fmt.Sprint(string(jsonData))
						// _, err = t.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
						// 	Peer:     &tg.InputPeerChannel{ChannelID: t.targetChatID, AccessHash: t.accessHash},
						// 	Message:  replyText,
						// 	ReplyTo:  &tg.InputReplyToMessage{ReplyToMsgID: msg.ID},
						// 	RandomID: int64(msg.ID), // Random ID unik untuk menghindari duplicate
						// })
						//
						// if err != nil {
						// 	log.Printf("❌ Gagal membalas pesan: %v", err)
						// } else {
						// 	log.Printf("✅ Berhasil membalas pesan ID %d dengan: %s", msg.ID, replyText)
						// }

					}
					// Tampilkan JSON
					// fmt.Println(string(jsonData))

				}
			}
		}
		return nil
	})
}

// ReplyToMessage -> Membalas pesan tertentu
func (t *TelegramClient) ReplyToMessage(ctx context.Context, messageID int, replyText string) {
	if t.targetChatID == 0 {
		log.Println("❌ Target Chat ID belum di-set. Cek apakah ResolveChannelID berhasil.")
		return
	}

	max := big.NewInt(1000) // Angka maksimal 100
	msgID, _ := rand.Int(rand.Reader, max)

	_, err := t.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer:     &tg.InputPeerChannel{ChannelID: t.targetChatID, AccessHash: t.accessHash},
		Message:  replyText,
		ReplyTo:  &tg.InputReplyToMessage{ReplyToMsgID: messageID},
		RandomID: msgID.Int64(), // Random ID unik untuk menghindari duplicate
	})

	if err != nil {
		log.Printf("❌ Gagal membalas pesan: %v", err)
	} else {
		log.Printf("✅ Berhasil membalas pesan ID %d dengan: %s", messageID, replyText)
	}
}

// Start -> Menjalankan Telegram Client
func (t *TelegramClient) Start() error {
	fmt.Println("✅ Telegram Client Berjalan...")
	return t.client.Run(ctx, func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})
}
