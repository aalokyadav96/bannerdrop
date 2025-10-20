package filedrop

import (
	"naevis/filemgr"
	"naevis/models"
	"naevis/mq"
)

// -------------------- Audio Processing --------------------

func processAudio(savedPath, uploadDir, uniqueID string, entitytype filemgr.EntityType) ([]int, []string) {
	_ = entitytype
	resolutions, outputPath := processAudioResolutions(savedPath, uploadDir, uniqueID)
	var paths []string
	if outputPath != "" {
		paths = []string{normalizePath(outputPath)}
	}

	go createSubtitleFile(uniqueID)
	mq.Notify("postaudio-uploaded", models.Index{})

	return resolutions, paths
}
