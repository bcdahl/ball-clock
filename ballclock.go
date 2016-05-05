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

// Given a number of balls display how many days until the balls return to their original order
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

// Run a standard clock for a specified time. Output the time and state of the clock (if specified)
func RunClock(timeToRun time.Duration, printStacks bool) {
	timerStop := time.NewTimer(timeToRun).C
	timerChan := time.NewTicker(time.Second * 60).C
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

// Given an initial count of balls run for the specified number of minutes
// When the time is reached output a json message that displays the current state of the clock
func RunForTime(balls int, minutes int) {
	if balls < 27 || balls > 127 {
		fmt.Println("Balls must be in the range of 27 - 127")
		return
	}
	if minutes < 0 {
		fmt.Println("Minutes must be > 0")
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

// Implements a Queue to maintain the main collection of balls (FIFO)
type BallQueue struct {
	head  int
	tail  int
	count int
	balls []ClockBall
}

// Push a ball on to the queue
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

// Pop a ball off the queue
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

// check if it is full
func (q *BallQueue) IsFull() bool {
	return q.count == len(q.balls)
}

// Check if it is empty
func (q *BallQueue) IsEmpty() bool {
	return q.count == 0
}

// How many balls are in the queue
func (q *BallQueue) Count() int {
	return q.count
}

// Determine if the order of the balls has returned back to its starting order
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

// Print the current state of the Queue
// [b1,b2,b3,b5,...]
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

// Maintain a Stack of balls (FILO)
type BallStack struct {
	head  int
	balls []ClockBall
}

// Push a ball on to the Stack
func (q *BallStack) Push(ball ClockBall) error {
	if q.IsFull() {
		return errors.New("Can't add any more")
	}
	q.balls[q.head] = ball
	q.head++
	return nil
}

// Pop a ball off of the Stack
func (q *BallStack) Pop() (ball ClockBall, err error) {
	if q.IsEmpty() {
		return ball, errors.New("Nothing to give you")
	}
	q.head--
	ball = q.balls[q.head]
	q.balls[q.head] = ClockBall{}
	return ball, nil
}

// Is the Stack full?
func (q *BallStack) IsFull() bool {
	return q.head == len(q.balls)
}

// Is the Stack Empty?
func (q *BallStack) IsEmpty() bool {
	return q.head == 0
}

// How many balls are on the Stack?
func (q *BallStack) Count() int {
	return q.head
}

// Create a string that represents the state of the Stack
// [b1,b2,b3,...]
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

// Maintain all of the necessary information about the state of the Clock
type BallClock struct {
	queue        BallQueue // Main collection of balls to use
	minute       BallStack // Indicate the number of minutes that have passed
	fiveMinute   BallStack // Indicate the number of Five minute durations that have passed
	hour         BallStack // Indicate the number of hours that have passed
	halfDays     int // maintain the number of full iterations of the Clock == 12 hours
	daysToRepeat int // Once determined, this will hold the number of days it took to repeat the main Queue of balls
	totalMinutes int // A running total of the elapsed minutes
	advanceTime  chan int // The channel to signal when the clock should advance
	done         chan int // The channel that a ball from the Queue should signal when it has reached its final position during a Tick
}

// A struct that represents a ball in transit
type ClockBallInTransit struct {
	ball ClockBall
	done chan int // The channel this Ball should signal when it is done moving. <nil> if it is not moving from the Queue
}

// Simple struct that holds the balls number
type ClockBall struct {
	number int
}

// The ball has reached its final stop, so signal the done channel if it is not <nil>
// Clear the channel once it has been signaled
func (b *ClockBallInTransit) Done() {
	if b.done != nil {
		c := b.done
		b.done = nil
		c <- 1
	}
}

// Perform a clock tick
// The tick is completed when the clock says it is
func (c *BallClock) Tick() {
	c.advanceTime <- 1
	<-c.done
}

// Initialize the Clocks queue with the specified number of balls
// Return the needed channels for the upstream states
// ballReturn channel is the channel that all states should use to return balls that are no longer needed
// ballAdvance channel is the channel that the Main state will use to send a ball to the next state
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

// Move the balls on one of the Stacks (Minute, FiveMinute, Hour)
// advance channel is the channel on which to listen for incoming balls from the previous state
// ballReturn channel is the channel to use to return balls that are no longer needed
// out channel is the channel to use to send a ball to the next state
func MoveBalls(s *BallStack, advance chan ClockBallInTransit, ballReturn chan ClockBallInTransit, out chan ClockBallInTransit) {
	for {
		select {
		case bit := <-advance:
			// Is this the last ball that will fit into this state?
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
				// Ball has reached the end of it's journey
				bit.Done()
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

// Initialize the Minute State
// minuteAdvance channel is the channel on which to listen for incoming balls from the previous state
// ballReturn channel is the channel to use to return balls that are no longer needed
// return a channel to use to send a ball to the next state
func (c *BallClock) InitMinute(minuteAdvance chan ClockBallInTransit, ballReturn chan ClockBallInTransit) chan ClockBallInTransit {
	c.minute = BallStack{balls: make([]ClockBall, 4)}
	out := make(chan ClockBallInTransit)
	go MoveBalls(&c.minute, minuteAdvance, ballReturn, out)
	return out
}

// Initialize the FiveMinute State
// fiveMinuteAdvance channel is the channel on which to listen for incoming balls from the previous state
// ballReturn channel is the channel to use to return balls that are no longer needed
// return a channel to use to send a ball to the next state
func (c *BallClock) InitFiveMinute(fiveMinuteAdvance chan ClockBallInTransit, ballReturn chan ClockBallInTransit) chan ClockBallInTransit {
	c.fiveMinute = BallStack{balls: make([]ClockBall, 11)}
	out := make(chan ClockBallInTransit)
	go MoveBalls(&c.fiveMinute, fiveMinuteAdvance, ballReturn, out)
	return out
}

// Initialize the Hour State
// hourAdvance channel is the channel on which to listen for incoming balls from the previous state
// ballReturn channel is the channel to use to return balls that are no longer needed
func (c *BallClock) InitHour(hourAdvance chan ClockBallInTransit, ballReturn chan ClockBallInTransit) {
	c.hour = BallStack{balls: make([]ClockBall, 11)}
	// Send the last ball back to the ballReturn no additional states
	go MoveBalls(&c.hour, hourAdvance, ballReturn, ballReturn)
}

// Create a ball clock and wire up the channels
func CreateBallClock(balls int) (clock *BallClock) {
	clock = &BallClock{advanceTime: make(chan int),
		done: make(chan int)}
	br, ma := clock.InitQueue(balls)
	fma := clock.InitMinute(ma, br)
	ha := clock.InitFiveMinute(fma, br)
	clock.InitHour(ha, br)
	return clock
}

// Has the ball clock returned to its original state?
func (c *BallClock) HasRepeated() bool {
	return c.queue.HasRepeated()
}

// How many minutes have passed?
func (c *BallClock) MinutesPassed() int {
	return c.totalMinutes
}

// Return a string that displays the current time of the clock
func (c *BallClock) Time() string {
	return fmt.Sprintf("Time is : %d:%02d", c.hour.Count()+1, (c.fiveMinute.Count()*5 + c.minute.Count()))
}

// Print the current state of the clock
// Min:[b1,b4]
// FiveMin:[b6,b10,b17,b2]
// Hour:[]
// Main:[b12,b19,b3,...]
func (c *BallClock) PrintTimeStacks() string {
	return fmt.Sprintf("Min:%v\nFiveMin:%v\nHour:%v\nMain:%v\n", c.minute, c.fiveMinute, c.hour, c.queue)
}
