package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	uploadpb "github.com/dimk00z/grpc-filetransfer/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientService struct {
	addr      string
	tls       bool
	filePath  string
	batchSize int
	client    uploadpb.FileServiceClient
	verbose   int
	testName  string
}

func New(addr string, tls bool, filePath string, batchSize int, verbose int, testName string) *ClientService {
	return &ClientService{
		addr:      addr,
		tls:       tls,
		filePath:  filePath,
		batchSize: batchSize,
		verbose:   verbose,
		testName:  testName,
	}
}

func (s *ClientService) SendFile() error {
	log.Println(s.addr, s.filePath)

	dialOptions := []grpc.DialOption{}
	if s.tls {
		log.Println("TLS enabled")
		// dail server
		config := &tls.Config{
			Certificates:       []tls.Certificate{},
			InsecureSkipVerify: true,
		}

		tlsCredential := credentials.NewTLS(config)
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(tlsCredential))
	} else {
		log.Println("TLS disabled")
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	//conn, err := grpc.Dial(s.addr, grpc.WithInsecure())
	conn, err := grpc.Dial(s.addr, dialOptions...)
	if err != nil {
		return err
	}
	defer conn.Close()
	s.client = uploadpb.NewFileServiceClient(conn)
	interrupt := make(chan os.Signal, 1)
	shutdownSignals := []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
	signal.Notify(interrupt, shutdownSignals...)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(s *ClientService) {
		if err = s.upload(ctx, cancel); err != nil {
			log.Fatal(err)
			cancel()
		}
	}(s)

	select {
	case killSignal := <-interrupt:
		log.Println("Got ", killSignal)
		cancel()
	case <-ctx.Done():
	}
	return nil
}

func (s *ClientService) upload(ctx context.Context, cancel context.CancelFunc) error {
	stream, err := s.client.Upload(ctx)
	if err != nil {
		return err
	}
	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	buf := make([]byte, s.batchSize)
	batchNumber := 1
	cps := map[int]int{}
	var idx int

	var testName string
	if s.testName != "" {
		testName = fmt.Sprintf("%s ", s.testName)
	}

	start := time.Now()

	for {
		num, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		chunk := buf[:num]

		if err := stream.Send(&uploadpb.FileUploadRequest{FileName: s.filePath, Chunk: chunk}); err != nil {
			return err
		}

		idx = int(time.Since(start).Seconds())

		if s.verbose > 0 && batchNumber > 0 && batchNumber%s.verbose == 0 {
			log.Printf("Sent %s- batch #%v(%d) - size %v\n", testName, batchNumber, idx, len(chunk))
		}

		batchNumber += 1

		cps[idx]++
	}

	cps[idx]++

	res, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	elapsed := time.Since(start)

	log.Printf("Sent %s- %s - %v bytes - %d chunks(%d) - %.3fs elapsed - %v chunkspersec - %s \n",
		testName,
		s.addr,
		res.GetSize(),
		batchNumber,
		s.batchSize,
		elapsed.Seconds(),
		batchNumber/int(elapsed.Seconds()),
		res.GetFileName())

	log.Printf("%sCPS: %+v \n", testName, cps)
	cancel()

	return nil
}
