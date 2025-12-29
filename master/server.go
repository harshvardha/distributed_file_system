package master

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/harshvardha/distributed_file_system/common"
	pb "github.com/harshvardha/distributed_file_system/proto"
	"google.golang.org/grpc"
)

// Server represents the master server
type Server struct {
	pb.UnimplementedMasterServer
	metadata *Metadata
	address  string
}

// NewServer creates a new master server
func NewServer(address string) *Server {
	return &Server{
		metadata: NewMetadata(),
		address:  address,
	}
}

// UploadFile handles file upload requests
func (s *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	log.Printf("Upload request for file: %s, size: %d bytes", req.Filename, req.Filesize)

	// Calculating number of chunks needed for storing the file
	numChunks := common.CalculateNumChunks(req.Filesize)

	// Adding file metadata
	s.metadata.AddFile(req.Filename, req.Filesize, numChunks)

	// Allocating chunks and assigning chunk servers
	chunkLocations := make([]*pb.ChunkLocation, 0, numChunks)

	for i := range numChunks {
		// Generating chunk handle for each chunk
		chunkHandle := common.GenerateChunkHandle(req.Filename, i)

		// Adding chunk metadata
		s.metadata.AddChunk(chunkHandle, req.Filename, int32(i))
		s.metadata.AddChunkToFile(req.Filename, chunkHandle)

		// fetching available chunk servers for replication
		servers := s.metadata.GetAvailableChunkServers(common.ReplicationFactor)

		if len(servers) < common.ReplicationFactor {
			log.Printf("Warning: Only %d chunk servers available, need %d for replication", len(servers), common.ReplicationFactor)
		}

		// Adding chunk location info
		chunkLocations = append(chunkLocations, &pb.ChunkLocation{
			ChunkHandle:          chunkHandle,
			ChunkServerAddresses: servers,
			ChunkIndex:           int32(i),
		})

		log.Printf("Chunk %d (%s) assigned to servers: %v", i, chunkHandle, servers)
	}

	return &pb.UploadFileResponse{
		ChunkLocations: chunkLocations,
	}, nil
}

// DownloadFile handles file download requests
func (s *Server) DownloadFile(ctx context.Context, req *pb.DownloadFileRequest) (*pb.DownloadFileResponse, error) {
	log.Printf("Download request for file: %s", req.Filename)

	// Get file metadata
	file, exists := s.metadata.GetFile(req.Filename)
	if !exists {
		return nil, fmt.Errorf("file not found: %s", req.Filename)
	}

	// Fetching chunk locations
	chunkLocations := make([]*pb.ChunkLocation, 0, len(file.Chunks))

	for _, chunkHandle := range file.Chunks {
		chunk, exists := s.metadata.GetChunk(chunkHandle)
		if !exists {
			return nil, fmt.Errorf("chunk not found: %s", chunkHandle)
		}

		chunkLocations = append(chunkLocations, &pb.ChunkLocation{
			ChunkHandle:          chunkHandle,
			ChunkServerAddresses: chunk.Locations,
			ChunkIndex:           chunk.ChunkIndex,
		})
	}

	return &pb.DownloadFileResponse{
		Filesize:      file.Filesize,
		ChunkLocation: chunkLocations,
	}, nil
}

// ListFiles handles list files request
func (s *Server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	log.Printf("List files request")

	files := s.metadata.ListFiles()
	fileInfos := make([]*pb.FileInfo, 0, len(files))

	for _, file := range files {
		fileInfos = append(fileInfos, &pb.FileInfo{
			Filename:  file.Filename,
			Filesize:  file.Filesize,
			NumChunks: int32(file.ChunkCount),
		})
	}

	return &pb.ListFilesResponse{
		Files: fileInfos,
	}, nil
}

// Heartbeat handles chunk server heartbeat
func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	log.Printf("Heartbeat from chunk server: %s with %d chunks", req.ChunkServerAddress, len(req.ChunkHandles))

	// registering/updating chunk server
	s.metadata.RegisterChunkServer(req.ChunkServerAddress, req.ChunkHandles)

	return &pb.HeartbeatResponse{
		Success: true,
	}, nil
}

// ReportChunk handles chunk storage completion reports
func (s *Server) ReportChunk(ctx context.Context, req *pb.ReportChunkRequest) (*pb.ReportChunkResponse, error) {
	log.Printf("Chunk report: %s stored on %s", req.ChunkHandle, req.ChunkServerAddress)

	// Adding chunk location
	s.metadata.AddChunkLocation(req.ChunkHandle, req.ChunkServerAddress)

	return &pb.ReportChunkResponse{
		Success: true,
	}, nil
}

// Start starts the master server
func (s *Server) Start() error {
	listen, err := net.Listen("tcp", common.MasterAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMasterServer(grpcServer, s)

	log.Printf("Master server starting on %s", s.address)

	if err := grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
