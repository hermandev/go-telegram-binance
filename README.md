# Bot Trade Coin Binance With Channel Telegram

ubah file `.env-example` menjadi `.env`

edit `TELE_API_ID` dan `TELE_API_HASH` dengann `APIID` dan API_HASH dari telegram telegram
kemudian buat channel di telegam dan edit `TELE_CHANNEL_NAME` dengan username channel telegram yang dibuat
edit `BINANCE_API_KEY` dan `BINANCE_SECRET_KEY` sesuai dengan yang akun binance anda

jalankan `go mod tidy` untuk menginstall dependency yang dibutuhkan
kemudian jalankan `go run main.go` untuk menjalankan program

setelah dijalankan akan meminta code verifikasi dari telegram anda..
