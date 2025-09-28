package main

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/yhonda-ohishi/etc_meisai_scraper/src/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// gRPCサーバーに接続
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewDownloadBufferServiceClient(conn)
	ctx := context.Background()

	// 方法1: CSVデータを直接バイナリで取得
	fmt.Println("=== Method 1: Direct Buffer ===")
	bufferResp, err := client.DownloadAsBuffer(ctx, &pb.BufferDownloadRequest{
		Accounts: []string{"account1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Received CSV data: %d bytes\n", bufferResp.SizeBytes)
	fmt.Printf("Record count: %d\n", bufferResp.RecordCount)
	// CSVデータを直接処理
	csvContent := string(bufferResp.CsvData)
	fmt.Printf("First 100 chars: %s...\n", csvContent[:min(100, len(csvContent))])

	// 方法2: ストリーミングで受信
	fmt.Println("\n=== Method 2: Streaming ===")
	stream, err := client.DownloadStream(ctx, &pb.BufferDownloadRequest{
		Accounts: []string{"account1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	})
	if err != nil {
		log.Fatal(err)
	}

	var fullData []byte
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fullData = append(fullData, chunk.Chunk...)
		fmt.Printf("Received chunk %d: %d bytes\n", chunk.SequenceNumber, len(chunk.Chunk))
	}
	fmt.Printf("Total received: %d bytes\n", len(fullData))

	// 方法3: 構造化データとして取得
	fmt.Println("\n=== Method 3: Structured Proto ===")
	protoResp, err := client.DownloadAsProto(ctx, &pb.BufferDownloadRequest{
		Accounts: []string{"account1"},
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total records: %d\n", protoResp.Metadata.TotalCount)
	for i, record := range protoResp.Records[:min(3, len(protoResp.Records))] {
		fmt.Printf("Record %d: %s %s %s→%s ¥%d\n",
			i+1, record.Date, record.Time,
			record.EntranceIc, record.ExitIc, record.Fare)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}