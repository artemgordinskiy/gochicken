package main

import "math"

// All generated sounds are raw stereo 16-bit little-endian PCM at sampleRate.
// They are passed directly to audioCtx.NewPlayerFromBytes.

const bytesPerSample = 4 // 2 channels × 2 bytes (16-bit)

func writeSample(buf []byte, i int, v float64) {
	if v > 1 {
		v = 1
	} else if v < -1 {
		v = -1
	}
	s := int16(v * 30000)
	off := i * bytesPerSample
	buf[off], buf[off+1] = byte(s), byte(s>>8)   // left
	buf[off+2], buf[off+3] = byte(s), byte(s>>8) // right
}

// genJumpSound is a rising chirp used as a fallback when jump.wav is missing.
func genJumpSound() []byte {
	dur := 0.14
	n := int(dur * sampleRate)
	buf := make([]byte, n*bytesPerSample)
	for i := range n {
		t := float64(i) / float64(sampleRate)
		freq := 380 + 900*(t/dur)
		env := math.Exp(-t * 8)
		writeSample(buf, i, env*math.Sin(2*math.Pi*freq*t))
	}
	return buf
}

// genHitSound is a descending tone mixed with noise — plays on death.
func genHitSound() []byte {
	dur := 0.45
	n := int(dur * sampleRate)
	buf := make([]byte, n*bytesPerSample)
	// deterministic pseudo-noise using a simple LCG so no rand import needed
	var seed uint64 = 12345
	for i := range n {
		seed = seed*6364136223846793005 + 1442695040888963407
		noise := float64(int64(seed>>33)) / float64(1<<31) // [-1, 1)
		t := float64(i) / float64(sampleRate)
		env := math.Exp(-t * 6)
		freq := 380 * math.Exp(-t*5) // sweep down
		tone := math.Sin(2 * math.Pi * freq * t)
		writeSample(buf, i, env*(tone*0.55+noise*0.45))
	}
	return buf
}

// genLandSound is a short low-frequency thud — plays when the chicken lands.
func genLandSound() []byte {
	dur := 0.10
	n := int(dur * sampleRate)
	buf := make([]byte, n*bytesPerSample)
	for i := range n {
		t := float64(i) / float64(sampleRate)
		env := math.Exp(-t * 38)
		writeSample(buf, i, env*math.Sin(2*math.Pi*85*t)*0.9)
	}
	return buf
}

// genMilestoneSound is three ascending tones — plays every 10 points.
func genMilestoneSound() []byte {
	freqs := []float64{550, 750, 1050}
	noteDur := 0.09
	n := int(float64(len(freqs)) * noteDur * float64(sampleRate))
	buf := make([]byte, n*bytesPerSample)
	samplesPerNote := int(noteDur * float64(sampleRate))
	for noteIdx, freq := range freqs {
		for j := range samplesPerNote {
			t := float64(j) / float64(sampleRate)
			env := math.Exp(-t * 18)
			i := noteIdx*samplesPerNote + j
			if i < n {
				writeSample(buf, i, env*math.Sin(2*math.Pi*freq*t)*0.75)
			}
		}
	}
	return buf
}
