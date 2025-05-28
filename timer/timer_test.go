package timer

import (
	"fmt"
	"testing"
	"time"
)

func TestCycleMode(t *testing.T) {
	timer := Timer{
		config: DefaultConfig,
  }
  timer.Init()
  for range 3 {
    timer.CycleMode()
    if timer.Mode != ShortBreak {
      t.Fatal("Failed cycle mode short break")
    }
    timer.CycleMode()
    if timer.Mode != Pomodoro {
      t.Fatal("Failed cycle mode to pomodoro")
    }
  }
  timer.CycleMode()
  if timer.Mode != LongBreak {
    t.Fatal("Failed cycle mode to long break", timer.Mode)
  }
}

func TestTick(t *testing.T) {
  var config = DefaultConfig
  config.DurationPerTick = time.Millisecond*10

	timer := Timer{
		config: config,
  }
  timer.Init()
  for i := range 5 {
    expected := config.Duration[Pomodoro]-time.Duration(i)*time.Millisecond*10
    if timer.Duration != expected {
      t.Fatalf("Failed loop %d != %d", timer.Duration, expected)
    }
    timer.tick()
  }
}

func TestLoop(t *testing.T) {
  var config = DefaultConfig
  config.DurationPerTick = time.Millisecond*10

	timer := Timer{
		config: config,
  }
  timer.Init()
  go timer.Loop()
  time.Sleep(time.Millisecond*21)
  expected := DefaultConfig.Duration[Pomodoro]-2*time.Millisecond*10
  if timer.Duration != expected {
    t.Fatalf("Failed loop %d != %d", timer.Duration, expected)
  }
}

func TestTimer(t *testing.T) {
  var config = DefaultConfig
  config.DurationPerTick = time.Millisecond*10
  config.SessionCount = 2
  config.Duration[Pomodoro] = 4*config.DurationPerTick
  config.Duration[ShortBreak] = 2*config.DurationPerTick
  config.Duration[LongBreak] = 5*config.DurationPerTick

	timer := Timer{
		config: config,
  }
  timer.Init()
  go timer.Loop()
  time.Sleep(time.Millisecond*45)
  if timer.Mode != ShortBreak {
    t.Fatalf("Didn't cycle mode to short break %s!=%s. time left %s", timer.Mode, ShortBreak, timer.String())
  }
  time.Sleep(time.Millisecond*20)
  if timer.Mode != Pomodoro {
    t.Fatalf("Didn't cycle mode to pomodoro %s!=%s. time left %s", timer.Mode, Pomodoro, timer.String())
  }
  time.Sleep(time.Millisecond*45)
  if timer.Mode != LongBreak {
    t.Fatalf("Didn't cycle mode to long break %s!=%s. time left %s", timer.Mode, LongBreak, timer.String())
  }
}

func ExampleTimer_String() {
	timer := Timer{
		config: DefaultConfig,
  }
  timer.Init()
  fmt.Println(timer.String())
  timer.config.Duration[Pomodoro] = 3 * time.Hour + 8 * time.Minute + 10 * time.Second
  timer.Init()
  fmt.Println(timer.String())
  timer.config.Duration[Pomodoro] = 50 * time.Second
  timer.Init()
  fmt.Println(timer.String())
  // Output:
  // 25m0s
  // 3h8m10s
  // 50s
}
