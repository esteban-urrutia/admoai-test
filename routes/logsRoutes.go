package routes

import (
    "archive/zip"
    "bytes"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

// parseDateWithFallback parses date string with multiple format support
func parseDateWithFallback(dateStr string, isEndDate bool) (time.Time, error) {
    loc := time.UTC
    fullLayout := "2006-01-02 15:04:05"
    dateLayout := "2006-01-02"

    // Try parsing with full datetime format first
    if parsedTime, err := time.ParseInLocation(fullLayout, dateStr, loc); err == nil {
        return parsedTime, nil
    }

    // If full datetime parsing fails, try date-only format
    parsedTime, err := time.ParseInLocation(dateLayout, dateStr, loc)
    if err != nil {
        return time.Time{}, fmt.Errorf("invalid date format. Use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS")
    }

    // If only date provided and it's endDate, set time to end of day
    if isEndDate {
        parsedTime = parsedTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
    }

    return parsedTime, nil
}

// validateDateRange validates that startDate is not after endDate
func validateDateRange(startDate, endDate time.Time) error {
    if startDate.After(endDate) {
        return fmt.Errorf("startDate cannot be after endDate")
    }
    return nil
}

// fetchDockerLogs executes docker logs command and returns the output
func fetchDockerLogs(containerName string, startDate, endDate time.Time) ([]byte, error) {
    sinceFormatted := startDate.Format(time.RFC3339)
    untilFormatted := endDate.Format(time.RFC3339)

    cmdArgs := []string{
        "logs",
        containerName,
        "--since", sinceFormatted,
        "--until", untilFormatted,
        "--timestamps",
    }

    cmd := exec.Command("docker", cmdArgs...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("error executing 'docker logs' on container '%s': %v. output: %s", containerName, err, string(output))
    }

    return output, nil
}

// createSafeFilename creates safe filename strings by replacing unsafe characters
func createSafeFilename(containerName, startDateStr, endDateStr string) (string, string) {
    safeContainerName := strings.ReplaceAll(containerName, ":", "_")
    safeContainerName = strings.ReplaceAll(safeContainerName, "/", "_")

    safeStartDate := strings.ReplaceAll(startDateStr, " ", "_")
    safeStartDate = strings.ReplaceAll(safeStartDate, ":", "-")
    safeEndDate := strings.ReplaceAll(endDateStr, " ", "_")
    safeEndDate = strings.ReplaceAll(safeEndDate, ":", "-")

    logFileName := fmt.Sprintf("logs_%s_%s_to_%s.log", safeContainerName, safeStartDate, safeEndDate)
    zipFileName := fmt.Sprintf("logs_%s_%s_to_%s.zip", safeContainerName, safeStartDate, safeEndDate)

    return logFileName, zipFileName
}

// createZipArchive creates a zip archive containing the log data
func createZipArchive(logData []byte, logFileName string) (*bytes.Buffer, error) {
    buf := new(bytes.Buffer)
    zipWriter := zip.NewWriter(buf)

    zipFile, err := zipWriter.Create(logFileName)
    if err != nil {
        return nil, fmt.Errorf("error creating zip file entry: %v", err)
    }

    _, err = zipFile.Write(logData)
    if err != nil {
        return nil, fmt.Errorf("error writing to zip file entry: %v", err)
    }

    err = zipWriter.Close()
    if err != nil {
        return nil, fmt.Errorf("error closing zip writer: %v", err)
    }

    return buf, nil
}

// Sets up the route for logs downloading
func Logs(router *gin.Engine) {
    router.GET("/logs", func(c *gin.Context) {
        startDateStr := c.Query("startDate")
        endDateStr := c.Query("endDate")
        containerName := "admoai-test-app-1" // Default container name

        // Validate required parameters
        if startDateStr == "" || endDateStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "startDate or endDate not found"})
            return
        }

        // Parse start date
        startDateTime, err := parseDateWithFallback(startDateStr, false)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate: " + err.Error()})
            return
        }

        // Parse end date
        endDateTime, err := parseDateWithFallback(endDateStr, true)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate: " + err.Error()})
            return
        }

        // Validate date range
        if err := validateDateRange(startDateTime, endDateTime); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Fetch docker logs
        output, err := fetchDockerLogs(containerName, startDateTime, endDateTime)
        if err != nil {
            log.Println(err.Error())
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
            return
        }

        // Check if logs are empty
        if len(output) == 0 {
            c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("No logs found on container '%s' between %s and %s", containerName, startDateStr, endDateStr)})
            return
        }

        // Create safe filenames
        logFileName, zipFileName := createSafeFilename(containerName, startDateStr, endDateStr)

        // Create zip archive
        buf, err := createZipArchive(output, logFileName)
        if err != nil {
            log.Printf("Error creating zip archive: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip archive"})
            return
        }

        // Set headers for file download
        c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipFileName))
        c.Header("Content-Type", "application/zip")

        // Send zip file as download
        c.Data(http.StatusOK, "application/zip", buf.Bytes())
    })
}