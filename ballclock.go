package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

func main() {
	//RunClock(time.Second*5, true)
	//FindDaysToRepeat(5)
	//FindDaysToRepeat(27)
	//FindDaysToRepeat(30)
	//FindDaysToRepeat(45)
	//FindDaysToRepeat(90)
	//FindDaysToRepeat(127)
	//FindDaysToRepeat(200)
	RunForTime(30, 325)
}

func FindDaysToRepeat(balls int) {
	if balls < 27 || balls > 127 {
		fmt.Println("Balls must be in the range of 27 - 127")
		return
	}
	clock := CreateBallClock(balls)
	for {
		clock.Tick()
		if clock.HasRepeated() {
			break
		}
	}
	fmt.Printf("%v balls cycle after %v days.\n", balls, clock.daysToRepeat)
}

func RunClock(timeToRun time.Duration, printStacks bool) {
	timerStop := time.NewTimer(timeToRun).C
	timerChan := time.NewTicker(time.Second).C
	clock := CreateBallClock(30)
	fmt.Println(clock.Time())
	for {
		select {
		case <-timerChan:
			clock.Tick()
			fmt.Println(clock.Time())
			if printStacks {
				fmt.Print(clock.PrintTimeStacks())
			}
		case <-timerStop:
			return
		}
	}
}

func RunForTime(balls int, minutes int) {
	if balls < 27 || balls > 127 {
		fmt.Println("Balls must be in the range of 27 - 127")
		return
	}
	clock := CreateBallClock(balls)
	for {
		clock.Tick()
		if clock.MinutesPassed() == minutes {
			break
		}
	}
	fmt.Printf("{\"Min\":%v,\"FiveMin\":%v,\"Hour\":%v,\"Main\":%v}", clock.minute, clock.fiveMinute, clock.hour, clock.queue)
}

type BallList interface {
	Push(ball ClockBall) error
	Pop() (ball ClockBall, err error)
	IsFull() bool
	IsEmpty() bool
	Count() int
}

type BallQueue struct {
	head  int
	tail  int
	count int
	balls []ClockBall
}

func (q *BallQueue) Push(ball ClockBall) error {
	if q.IsFull() {
		return errors.New("Can't add any more")
	}
	q.balls[q.head] = ball
	q.head++
	if q.head == len(q.balls) {
		q.head = 0
	}
	q.count++
	return nil
}

func (q *BallQueue) Pop() (ball ClockBall, err error) {
	if q.IsEmpty() {
		return ball, errors.New("Nothing to give you")
	}
	ball = q.balls[q.tail]
	q.balls[q.tail] = ClockBall{0}
	q.tail++
	if q.tail == len(q.balls) {
		q.tail = 0
	}
	q.count--
	return ball, nil
}

func (q *BallQueue) IsFull() bool {
	return q.count == len(q.balls)
}

func (q *BallQueue) IsEmpty() bool {
	return q.count == 0
}

func (q *BallQueue) Count() int {
	return q.count
}

func (q *BallQueue) HasRepeated() bool {
	if !q.IsFull() {
		return false
	}
	ordered := true
	index := q.tail
	n1 := q.balls[index].number
	for i := 1; i < len(q.balls); i++ {
		index++
		if index == len(q.balls) {
			index = 0
		}
		n2 := q.balls[index].number
		if n2-n1 != 1 {
			ordered = false
			break
		}
		n1 = n2
	}
	return ordered
}

func (q BallQueue) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	if !q.IsEmpty() {
		index := q.tail
		for i := 0; i < q.Count(); i++ {
			if index == len(q.balls) {
				index = 0
			}
			if i > 0 {
				buffer.WriteString(",")
			}
			buffer.WriteString(strconv.Itoa(q.balls[index].number))
			index++
		}
	}
	buffer.WriteString("]")
	return buffer.String()
}

type BallStack struct {
	head  int
	balls []ClockBall
}

func (q *BallStack) Push(ball ClockBall) error {
	if q.IsFull() {
		return errors.New("Can't add any more")
	}
	q.balls[q.head] = ball
	q.head++
	return nil
}

func (q *BallStack) Pop() (ball ClockBall, err error) {
	if q.IsEmpty() {
		return ball, errors.New("Nothing to give you")
	}
	q.head--
	ball = q.balls[q.head]
	q.balls[q.head] = ClockBall{}
	return ball, nil
}

func (q *BallStack) IsFull() bool {
	return q.head == len(q.balls)
}

func (q *BallStack) IsEmpty() bool {
	return q.head == 0
}

func (q *BallStack) Count() int {
	return q.head
}

func (q BallStack) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	for i := 0; i < q.head; i++ {
		if i > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(strconv.Itoa(q.balls[i].number))
	}
	buffer.WriteString("]")
	return buffer.String()
}

type BallClock struct {
	queue        BallQueue
	minute       BallStack
	fiveMinute   BallStack
	hour         BallStack
	halfDays     int
	daysToRepeat int
	totalMinutes int
	advanceTime  chan int
	done         chan int
}

type ClockBallInTransit struct {
	ball ClockBall
	done chan int
}

type ClockBall struct {
	number int
}

func (b *ClockBallInTransit) Done() {
	if b.done != nil {
		c := b.done
		b.done = nil
		c <- 1
	}
}

func (c *BallClock) Tick() {
	c.advanceTime <- 1
	<-c.done
}

func (c *BallClock) InitQueue(balls int) (ballReturn chan ClockBallInTransit, ballAdvance chan ClockBallInTransit) {
	c.queue = BallQueue{
		balls: make([]ClockBall, balls)}
	c.daysToRepeat = 0
	for i := 0; i < balls; i++ {
		c.queue.Push(ClockBall{i + 1})
	}
	ballReturn = make(chan ClockBallInTransit)
	ballAdvance = make(chan ClockBallInTransit)
	go func() {
		for {
			select {
			case bit := <-ballReturn:
				//Put the ball back in the queue
				c.queue.Push(bit.ball)
				// check if a half day has passed
				if c.queue.IsFull() {
					c.halfDays++
					// check if a full rotation has occurred
					if c.queue.HasRepeated() {
						c.daysToRepeat = c.halfDays / 2
					}
				}
				// Ball has potentially reached the end of it's jorney
				bit.Done()
			case <-c.advanceTime:
				// Send a ball on its way
				b, err := c.queue.Pop()
				if err != nil {
					fmt.Println(err)
					return
				}
				bit := ClockBallInTransit{b, c.done}
				ballAdvance <- bit
				c.totalMinutes++
			}
		}
	}()
	return ballReturn, ballAdvance
}

func MoveBalls(s *BallStack, advance chan ClockBallInTransit, ballReturn chan ClockBallInTransit, out chan ClockBallInTransit) {
	for {
		select {
		case bit := <-advance:
			// Is this the last ball?
			if s.IsFull() {
				// Return all the balls
				for {
					if s.IsEmpty() {
						break
					}
					br, err := s.Pop()
					if err != nil {
						fmt.Println(err)
						return
					}
					// Return the ball
					rbit := ClockBallInTransit{ball: br}
					ballReturn <- rbit
				}
				// Now send the last ball on
				out <- bit
			} else {
				// Just save it
				err := s.Push(bit.ball)
				// Ball has reached the end of it's jorney
				bit.Done()
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (c *BallClock) InitMinute(minuteAdvance chan ClockBallInTransit, ballReturn chan ClockBallInTransit) chan ClockBallInTransit {
	c.minute = BallStack{balls: make([]ClockBall, 4)}
	out := make(chan ClockBallInTransit)
	go MoveBalls(&c.minute, minuteAdvance, ballReturn, out)
	return out
}

func (c *BallClock) InitFiveMinute(fiveMinuteAdvance chan ClockBallInTransit, ballReturn chan ClockBallInTransit) chan ClockBallInTransit {
	c.fiveMinute = BallStack{balls: make([]ClockBall, 11)}
	out := make(chan ClockBallInTransit)
	go MoveBalls(&c.fiveMinute, fiveMinuteAdvance, ballReturn, out)
	return out
}

func (c *BallClock) InitHour(hourAdvance chan ClockBallInTransit, ballReturn chan ClockBallInTransit) {
	c.hour = BallStack{balls: make([]ClockBall, 11)}
	// Send the last ball back to the ballReturn no additional stages
	go MoveBalls(&c.hour, hourAdvance, ballReturn, ballReturn)
}

func CreateBallClock(balls int) (clock *BallClock) {
	clock = &BallClock{advanceTime: make(chan int),
		done: make(chan int)}
	br, ma := clock.InitQueue(balls)
	fma := clock.InitMinute(ma, br)
	ha := clock.InitFiveMinute(fma, br)
	clock.InitHour(ha, br)
	return clock
}

func (c *BallClock) HasRepeated() bool {
	return c.queue.HasRepeated()
}

func (c *BallClock) MinutesPassed() int {
	return c.totalMinutes
}

func (c *BallClock) Time() string {
	return fmt.Sprintf("Time is : %d:%02d", c.hour.Count()+1, (c.fiveMinute.Count()*5 + c.minute.Count()))
}

func (c *BallClock) PrintTimeStacks() string {
	return fmt.Sprintf("Min:%v\nFiveMin:%v\nHour:%v\nMain:%v\n", c.minute, c.fiveMinute, c.hour, c.queue)
}
