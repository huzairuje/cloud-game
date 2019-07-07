package nanoarch

import (
	"image"
)

/*
#include "libretro.h"
#cgo LDFLAGS: -ldl
#include <stdlib.h>
#include <stdio.h>
#include <dlfcn.h>
#include <string.h>

void bridge_retro_init(void *f);
void bridge_retro_deinit(void *f);
unsigned bridge_retro_api_version(void *f);
void bridge_retro_get_system_info(void *f, struct retro_system_info *si);
void bridge_retro_get_system_av_info(void *f, struct retro_system_av_info *si);
bool bridge_retro_set_environment(void *f, void *callback);
void bridge_retro_set_video_refresh(void *f, void *callback);
void bridge_retro_set_input_poll(void *f, void *callback);
void bridge_retro_set_input_state(void *f, void *callback);
void bridge_retro_set_audio_sample(void *f, void *callback);
void bridge_retro_set_audio_sample_batch(void *f, void *callback);
bool bridge_retro_load_game(void *f, struct retro_game_info *gi);
void bridge_retro_unload_game(void *f);
void bridge_retro_run(void *f);

bool coreEnvironment_cgo(unsigned cmd, void *data);
void coreVideoRefresh_cgo(void *data, unsigned width, unsigned height, size_t pitch);
void coreInputPoll_cgo();
void coreAudioSample_cgo(int16_t left, int16_t right);
size_t coreAudioSampleBatch_cgo(const int16_t *data, size_t frames);
int16_t coreInputState_cgo(unsigned port, unsigned device, unsigned index, unsigned id);
void coreLog_cgo(enum retro_log_level level, const char *msg);
*/
import "C"

// naEmulator implements CloudEmulator
type naEmulator struct {
	imageChannel chan<- *image.RGBA
	audioChannel chan<- float32
	inputChannel <-chan int
	corePath     string
	gamePath     string
	roomID       string

	keys []bool
}

var NAEmulator *naEmulator
var bindRetroKeys = map[int]int{
	0: C.RETRO_DEVICE_ID_JOYPAD_A,
	1: C.RETRO_DEVICE_ID_JOYPAD_B,
	2: C.RETRO_DEVICE_ID_JOYPAD_SELECT,
	3: C.RETRO_DEVICE_ID_JOYPAD_START,
	4: C.RETRO_DEVICE_ID_JOYPAD_UP,
	5: C.RETRO_DEVICE_ID_JOYPAD_DOWN,
	6: C.RETRO_DEVICE_ID_JOYPAD_LEFT,
	7: C.RETRO_DEVICE_ID_JOYPAD_RIGHT,
}

func NewNAEmulator(imageChannel chan<- *image.RGBA, inputChannel <-chan int) *naEmulator {
	return &naEmulator{
		//corePath:     "libretro/cores/pcsx_rearmed_libretro.so",
		corePath:     "libretro/cores/mgba_libretro.so",
		imageChannel: imageChannel,
		inputChannel: inputChannel,
		keys:         make([]bool, C.RETRO_DEVICE_ID_JOYPAD_R3+1),
	}
}

func Init(imageChannel chan<- *image.RGBA, inputChannel <-chan int) {
	NAEmulator = NewNAEmulator(imageChannel, inputChannel)
	go NAEmulator.listenInput()
}

func (na *naEmulator) listenInput() {
	// input from javascript follows bitmap. Ex: 00110101
	// we decode the bitmap and send to channel
	for inpBitmap := range NAEmulator.inputChannel {
		for k := 0; k < len(na.keys); k++ {
			if (inpBitmap & 1) == 1 {
				key := bindRetroKeys[k]
				na.keys[key] = true
			}
			inpBitmap >>= 1
		}
	}
}

func (na *naEmulator) Start(path string) {
	coreLoad(na.corePath)
	na.playGame(path)

	for {
		C.bridge_retro_run(retroRun)
	}
}

func (na *naEmulator) playGame(path string) {
	coreLoadGame(path)
}

func (na *naEmulator) SaveGame(saveExtraFunc func() error) error {
	return nil
}

func (na *naEmulator) LoadGame() error {
	return nil
}

func (na *naEmulator) GetHashPath() string {
	return savePath(na.roomID)
}

func savePath(hash string) string {
	//return homeDir + "/.nes/save/" + hash + ".dat"
	return ""
}

func (na *naEmulator) Close() {
	// Unload and deinit in the core.
	C.bridge_retro_unload_game(retroUnloadGame)
	C.bridge_retro_deinit(retroDeinit)
}