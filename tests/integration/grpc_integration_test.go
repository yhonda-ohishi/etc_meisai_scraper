package integration_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupTestServer() (*grpc.Server, error) {
	// Create test configuration
	// Note: Config types are not yet defined, using simplified setup

	// Initialize repositories - using mock implementations for now
	// TODO: Replace with actual repository implementations when available
	// etcRepo := repositories.NewInMemoryETCRepository()
	// mappingRepo := repositories.NewInMemoryMappingRepository()

	// Initialize services - temporarily skip until repositories are available
	// etcService := services.NewETCService(etcRepo)
	// mappingService := services.NewMappingService(mappingRepo, etcRepo)
	// importService := services.NewImportService(etcRepo, mappingRepo)

	// Create gRPC server
	s := grpc.NewServer()

	// Register service - temporarily skip until services are available
	// etcMeisaiServer := etcgrpc.NewETCMeisaiServer(etcService, mappingService, importService)
	// pb.RegisterETCMeisaiServiceServer(s, etcMeisaiServer)

	return s, nil
}

func TestGRPCIntegration_ETCMeisaiCRUD(t *testing.T) {
	grpcServer, err := setupTestServer()
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// Create client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)

	t.Run("CreateETCMeisai", func(t *testing.T) {
		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				EtcNum:        &[]string{"12345"}[0],
				Date:          "2024-01-01",
				Time:          "10:30",
				EntranceIc:    "東京IC",
				ExitIc:        "大阪IC",
				TollAmount:    1000,
				CarNumber:     "普通車",
				EtcCardNumber: "一般",
			},
		}

		resp, err := client.CreateRecord(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Record.Id)
		assert.Equal(t, "12345", *resp.Record.EtcNum)
	})

	t.Run("GetETCMeisai", func(t *testing.T) {
		// First create an ETC record
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				EtcNum:     &[]string{"67890"}[0],
				Date:       "2024-01-02",
				Time:       "15:45",
				EntranceIc: "名古屋IC",
				ExitIc:     "京都IC",
				TollAmount: 800,
			},
		}

		createResp, err := client.CreateRecord(context.Background(), createReq)
		require.NoError(t, err)

		// Now get the record
		getReq := &pb.GetRecordRequest{
			Id: createResp.Record.Id,
		}

		getResp, err := client.GetRecord(context.Background(), getReq)
		assert.NoError(t, err)
		assert.NotNil(t, getResp)
		assert.Equal(t, createResp.Record.Id, getResp.Record.Id)
		assert.Equal(t, "67890", *getResp.Record.EtcNum)
	})

	t.Run("ListETCMeisai", func(t *testing.T) {
		// Create multiple records
		for i := 0; i < 3; i++ {
			createReq := &pb.CreateRecordRequest{
				Record: &pb.ETCMeisaiRecord{
					EtcNum:     &[]string{fmt.Sprintf("LIST%d", i)}[0],
					Date:       "2024-01-03",
					Time:       "12:00",
					EntranceIc: "テストIC",
					ExitIc:     "テスト出口IC",
					TollAmount: int32(500 + i*100),
				},
			}
			_, err := client.CreateRecord(context.Background(), createReq)
			require.NoError(t, err)
		}

		// List records
		listReq := &pb.ListRecordsRequest{
			PageSize: 10,
		}

		listResp, err := client.ListRecords(context.Background(), listReq)
		assert.NoError(t, err)
		assert.NotNil(t, listResp)
		assert.GreaterOrEqual(t, len(listResp.Records), 3)
	})

	t.Run("UpdateETCMeisai", func(t *testing.T) {
		// Create a record to update
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				EtcNum:     &[]string{"UPDATE123"}[0],
				Date:       "2024-01-04",
				Time:       "08:00",
				EntranceIc: "更新前IC",
				ExitIc:     "更新前出口IC",
				TollAmount: 600,
			},
		}

		createResp, err := client.CreateRecord(context.Background(), createReq)
		require.NoError(t, err)

		// Update the record
		updateReq := &pb.UpdateRecordRequest{
			Id: createResp.Record.Id,
			Record: &pb.ETCMeisaiRecord{
				Id:         createResp.Record.Id,
				EtcNum:     &[]string{"UPDATE123"}[0],
				Date:       "2024-01-04",
				Time:       "08:00",
				EntranceIc: "更新後IC",
				ExitIc:     "更新後出口IC",
				TollAmount: 1200,
			},
		}

		updateResp, err := client.UpdateRecord(context.Background(), updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, updateResp)
		assert.Equal(t, "更新後IC", updateResp.Record.EntranceIc)
		assert.Equal(t, int32(1200), updateResp.Record.TollAmount)
	})

	t.Run("DeleteETCMeisai", func(t *testing.T) {
		// Create a record to delete
		createReq := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				EtcNum:     &[]string{"DELETE123"}[0],
				Date:       "2024-01-05",
				Time:       "20:00",
				EntranceIc: "削除IC",
				ExitIc:     "削除出口IC",
				TollAmount: 300,
			},
		}

		createResp, err := client.CreateRecord(context.Background(), createReq)
		require.NoError(t, err)

		// Delete the record
		deleteReq := &pb.DeleteRecordRequest{
			Id: createResp.Record.Id,
		}

		_, err = client.DeleteRecord(context.Background(), deleteReq)
		assert.NoError(t, err)

		// Verify deletion
		getReq := &pb.GetRecordRequest{
			Id: createResp.Record.Id,
		}

		_, err = client.GetRecord(context.Background(), getReq)
		assert.Error(t, err)
	})
}

func TestGRPCIntegration_ETCMappingOperations(t *testing.T) {
	t.Skip("Mapping operations test disabled - protobuf schema mismatch")
	/*
	grpcServer, err := setupTestServer()
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// Create client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)

	t.Run("CreateMapping", func(t *testing.T) {
		req := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcRecordId:      12345,
				MappingType:      "manual",
				MappedEntityId:   67890,
				MappedEntityType: "dtako_record",
				Confidence:       0.95,
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:        "test-user",
			},
		}

		resp, err := client.CreateMapping(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Mapping.Id)
		assert.Equal(t, int64(12345), resp.Mapping.EtcRecordId)
		assert.Equal(t, "manual", resp.Mapping.MappingType)
	})

	t.Run("AutoMatchMappings", func(t *testing.T) {
		// Create some ETC records first
		for i := 0; i < 3; i++ {
			createEtcReq := &pb.CreateETCMeisaiRequest{
				EtcMeisai: &pb.ETCMeisai{
					EtcNum:      fmt.Sprintf("AUTO%d", i),
					UseDate:     "2024-01-06",
					UseTime:     fmt.Sprintf("%02d:00", 10+i),
					InIcName:    "自動IC",
					OutIcName:   "自動出口IC",
					HighwayName: "自動高速",
					Amount:      int32(1000 + i*200),
				},
			}
			_, err := client.CreateETCMeisai(context.Background(), createEtcReq)
			require.NoError(t, err)
		}

		// Perform auto-matching
		autoMatchReq := &pb.AutoMatchRequest{
			DateRange: &pb.DateRange{
				StartDate: "2024-01-06",
				EndDate:   "2024-01-06",
			},
			MinMatchScore: 80,
		}

		autoMatchResp, err := client.AutoMatchMappings(context.Background(), autoMatchReq)
		assert.NoError(t, err)
		assert.NotNil(t, autoMatchResp)
		assert.GreaterOrEqual(t, autoMatchResp.MatchedCount, int32(0))
	})

	t.Run("ConfirmMapping", func(t *testing.T) {
		// Create a mapping first
		createReq := &pb.CreateMappingRequest{
			Mapping: &pb.ETCMapping{
				EtcNum:      "CONFIRM001",
				DtakoRowId:  67890,
				UseDate:     "2024-01-07",
				UseTime:     "14:00",
				InIcName:    "確認IC",
				OutIcName:   "確認出口IC",
				HighwayName: "確認高速",
				Amount:      2000,
				MatchScore:  90,
				IsConfirmed: false,
			},
		}

		createResp, err := client.CreateMapping(context.Background(), createReq)
		require.NoError(t, err)

		// Confirm the mapping
		confirmReq := &pb.ConfirmMappingRequest{
			MappingId: createResp.Mapping.Id,
		}

		confirmResp, err := client.ConfirmMapping(context.Background(), confirmReq)
		assert.NoError(t, err)
		assert.NotNil(t, confirmResp)
		assert.True(t, confirmResp.Mapping.IsConfirmed)
	})

	t.Run("GetMappingStatistics", func(t *testing.T) {
		req := &pb.GetMappingStatisticsRequest{
			DateRange: &pb.DateRange{
				StartDate: "2024-01-01",
				EndDate:   "2024-01-31",
			},
		}

		resp, err := client.GetMappingStatistics(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.Statistics.TotalRecords, int32(0))
		assert.GreaterOrEqual(t, resp.Statistics.MappedRecords, int32(0))
		assert.GreaterOrEqual(t, resp.Statistics.ConfirmedMappings, int32(0))
	})
	*/
}

func TestGRPCIntegration_ImportOperations(t *testing.T) {
	t.Skip("Import operations test disabled - protobuf schema mismatch")
	/*
	grpcServer, err := setupTestServer()
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// Create client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)

	t.Run("CreateImportSession", func(t *testing.T) {
		req := &pb.CreateImportSessionRequest{
			Filename: "test_import.csv",
			FileSize: 1024,
			FileHash: "abc123def456",
		}

		resp, err := client.CreateImportSession(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Session.Id)
		assert.Equal(t, "test_import.csv", resp.Session.Filename)
		assert.Equal(t, "pending", resp.Session.Status)
	})

	t.Run("ImportCSVData", func(t *testing.T) {
		// Create import session first
		sessionReq := &pb.CreateImportSessionRequest{
			Filename: "import_test.csv",
			FileSize: 2048,
			FileHash: "import123hash",
		}

		sessionResp, err := client.CreateImportSession(context.Background(), sessionReq)
		require.NoError(t, err)

		// Import CSV data
		csvData := `利用日,利用時刻,入口IC,出口IC,高速道路名,金額,車種,利用タイプ,ETC番号
2024-01-08,09:00,CSV入口IC,CSV出口IC,CSV高速,1500,普通車,一般,CSV001
2024-01-08,10:30,CSV入口IC2,CSV出口IC2,CSV高速2,2000,普通車,一般,CSV002`

		importReq := &pb.ImportCSVDataRequest{
			SessionId: sessionResp.Session.Id,
			CsvData:   csvData,
		}

		importResp, err := client.ImportCSVData(context.Background(), importReq)
		assert.NoError(t, err)
		assert.NotNil(t, importResp)
		assert.Equal(t, int32(2), importResp.ProcessedRecords)
		assert.Equal(t, int32(2), importResp.ImportedRecords)
		assert.Equal(t, int32(0), importResp.ErrorCount)
	})

	t.Run("GetImportSession", func(t *testing.T) {
		// Create session
		createReq := &pb.CreateImportSessionRequest{
			Filename: "get_session_test.csv",
			FileSize: 512,
			FileHash: "getsession123",
		}

		createResp, err := client.CreateImportSession(context.Background(), createReq)
		require.NoError(t, err)

		// Get session
		getReq := &pb.GetImportSessionRequest{
			SessionId: createResp.Session.Id,
		}

		getResp, err := client.GetImportSession(context.Background(), getReq)
		assert.NoError(t, err)
		assert.NotNil(t, getResp)
		assert.Equal(t, createResp.Session.Id, getResp.Session.Id)
		assert.Equal(t, "get_session_test.csv", getResp.Session.Filename)
	})

	t.Run("ListImportSessions", func(t *testing.T) {
		// Create multiple sessions
		for i := 0; i < 3; i++ {
			createReq := &pb.CreateImportSessionRequest{
				Filename: fmt.Sprintf("list_test_%d.csv", i),
				FileSize: int64(100 + i*50),
				FileHash: fmt.Sprintf("listhash%d", i),
			}
			_, err := client.CreateImportSession(context.Background(), createReq)
			require.NoError(t, err)
		}

		// List sessions
		listReq := &pb.ListImportSessionsRequest{
			PageSize:  10,
			PageToken: "",
		}

		listResp, err := client.ListImportSessions(context.Background(), listReq)
		assert.NoError(t, err)
		assert.NotNil(t, listResp)
		assert.GreaterOrEqual(t, len(listResp.Sessions), 3)
	})
	*/
}

func TestGRPCIntegration_StreamingOperations(t *testing.T) {
	t.Skip("Streaming operations test disabled - protobuf schema mismatch")
	/*
	grpcServer, err := setupTestServer()
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// Create client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)

	t.Run("StreamETCMeisai", func(t *testing.T) {
		// Create some records first
		for i := 0; i < 5; i++ {
			createReq := &pb.CreateETCMeisaiRequest{
				EtcMeisai: &pb.ETCMeisai{
					EtcNum:      fmt.Sprintf("STREAM%d", i),
					UseDate:     "2024-01-09",
					UseTime:     fmt.Sprintf("%02d:00", 8+i),
					InIcName:    "ストリームIC",
					OutIcName:   "ストリーム出口IC",
					HighwayName: "ストリーム高速",
					Amount:      int32(800 + i*100),
				},
			}
			_, err := client.CreateETCMeisai(context.Background(), createReq)
			require.NoError(t, err)
		}

		// Stream records
		req := &pb.StreamETCMeisaiRequest{
			DateRange: &pb.DateRange{
				StartDate: "2024-01-09",
				EndDate:   "2024-01-09",
			},
			BatchSize: 2,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		stream, err := client.StreamETCMeisai(ctx, req)
		require.NoError(t, err)

		var receivedRecords []*pb.ETCMeisai
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)

			receivedRecords = append(receivedRecords, resp.EtcMeisais...)
		}

		assert.GreaterOrEqual(t, len(receivedRecords), 5)
	})

	t.Run("WatchImportProgress", func(t *testing.T) {
		// Create import session
		sessionReq := &pb.CreateImportSessionRequest{
			Filename: "watch_progress_test.csv",
			FileSize: 4096,
			FileHash: "watchprogress123",
		}

		sessionResp, err := client.CreateImportSession(context.Background(), sessionReq)
		require.NoError(t, err)

		// Watch progress
		watchReq := &pb.WatchImportProgressRequest{
			SessionId: sessionResp.Session.Id,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		stream, err := client.WatchImportProgress(ctx, watchReq)
		require.NoError(t, err)

		// Start import in parallel
		go func() {
			time.Sleep(100 * time.Millisecond) // Give watch time to start

			csvData := `利用日,利用時刻,入口IC,出口IC,高速道路名,金額,車種,利用タイプ,ETC番号
2024-01-10,11:00,プログレスIC,プログレス出口IC,プログレス高速,3000,普通車,一般,PROG001`

			importReq := &pb.ImportCSVDataRequest{
				SessionId: sessionResp.Session.Id,
				CsvData:   csvData,
			}

			client.ImportCSVData(context.Background(), importReq)
		}()

		// Receive progress updates
		var progressUpdates []*pb.ImportProgress
		for i := 0; i < 3; i++ { // Limit to avoid infinite loop
			progress, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			progressUpdates = append(progressUpdates, progress)
		}

		assert.GreaterOrEqual(t, len(progressUpdates), 1)
	})
	*/
}

func TestGRPCIntegration_ErrorHandling(t *testing.T) {
	grpcServer, err := setupTestServer()
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	// Create client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewETCMeisaiServiceClient(conn)

	t.Run("GetNonExistentRecord", func(t *testing.T) {
		req := &pb.GetRecordRequest{
			Id: 99999, // non-existent ID
		}

		_, err := client.GetRecord(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("InvalidCreateRequest", func(t *testing.T) {
		req := &pb.CreateRecordRequest{
			Record: &pb.ETCMeisaiRecord{
				// Missing required fields
				EtcNum: &[]string{""}[0],
			},
		}

		_, err := client.CreateRecord(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req := &pb.ListRecordsRequest{
			PageSize: 100,
		}

		_, err := client.ListRecords(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deadline exceeded")
	})
}

