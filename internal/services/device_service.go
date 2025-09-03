package services

import (
	"fmt"

	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/models"
)

type DeviceService struct {
	config *config.Config
}

func NewDeviceService(cfg *config.Config) *DeviceService {
	return &DeviceService{
		config: cfg,
	}
}

// ValidateDevice validates a device by serial number
func (s *DeviceService) ValidateDevice(req *models.DeviceValidationRequest) (*models.DeviceValidationResponse, error) {
	if !s.config.DeviceAuthEnabled {
		// If device auth is disabled, allow all devices
		return &models.DeviceValidationResponse{
			Valid:    true,
			DeviceID: req.SerialNumber,
			Message:  "Device authentication disabled",
		}, nil
	}

	// Basic validation - in production, you'd validate against your device database
	if req.SerialNumber == "" {
		return &models.DeviceValidationResponse{
			Valid:   false,
			Message: "Serial number required",
		}, nil
	}

	// For now, accept any non-empty serial number
	// TODO: Implement actual device validation logic
	return &models.DeviceValidationResponse{
		Valid:    true,
		DeviceID: req.SerialNumber,
		Message:  "Device validated successfully",
	}, nil
}

// IsValidDevice checks if a device serial number is valid
func (s *DeviceService) IsValidDevice(serialNumber string) bool {
	req := &models.DeviceValidationRequest{
		SerialNumber: serialNumber,
	}
	
	resp, err := s.ValidateDevice(req)
	if err != nil {
		return false
	}
	
	return resp.Valid
}
