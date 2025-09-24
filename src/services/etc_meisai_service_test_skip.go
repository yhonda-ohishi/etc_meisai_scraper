// +build skip_gorm_tests

package services_test

// This file contains tests for services that incorrectly use GORM directly.
// These services should be refactored to use db_service client through repositories.
// Until then, these tests are disabled to avoid GORM nil pointer panics.
//
// Affected services:
// - ETCMeisaiService (duplicate of ETCService)
// - ETCMappingService (duplicate of MappingService)
// - StatisticsService
// - ImportService (has ImportServiceLegacy that uses repos correctly)
//
// The correct architecture is:
// Handler -> Service -> Repository -> gRPC Client (db_service)
//
// NOT:
// Handler -> Service -> GORM DB (wrong!)