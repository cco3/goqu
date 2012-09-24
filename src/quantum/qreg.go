// Copyright 2011 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//      Unless required by applicable law or agreed to in writing, software
//      distributed under the License is distributed on an "AS IS" BASIS,
//      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//      See the License for the specific language governing permissions and
//      limitations under the License.
//
// Author: conleyo@google.com (Conley Owens)

package quantum

import (
	"cmath"
	"fmt"
	"math"
	"rand"
	"time"
)

func init() {
	rand.Seed(time.Nanoseconds())
}

// Represents a quantum register
type QReg struct {
	states []complex128
	size   int // This is ln(len(states))
}

// Constructor for a QReg
func NewQReg(size int, value int) *QReg {
	qreg := &QReg{make([]complex128, 1<<uint(size)), size}
	qreg.Set(value)
	return qreg
}

// Accessor for the size of a QReg
func (qreg *QReg) Size() int {
	return qreg.size
}

// Copy a quantum register
func (qreg *QReg) Copy() *QReg {
	new_qreg := &QReg{make([]complex128, len(qreg.states)), qreg.size}
	copy(new_qreg.states, qreg.states)
	return new_qreg
}

// Get the probability of observing a state
func (qreg *QReg) StateProb(state int) float64 {
	return cmath.Abs(qreg.states[state] * qreg.states[state])
}

// Get the probability of observing a state for a specific bit
func (qreg *QReg) BProb(index int, value int) float64 {
	prob := float64(0.0)
	bit := 1 << uint(index)
	bitnot := (1 - value) << uint(index)
	// Iterate through all the states where this bit is 1
	for state := 0 | bit; state < len(qreg.states); state = (state + 1) | bit {
		prob += qreg.StateProb(state - bitnot)
	}
	return prob
}

// Set the state of the QReg
func (qreg *QReg) Set(value int) {
	if value > 1<<uint(qreg.size) {
		err_str := fmt.Sprintf("Value of %d is too large for QReg "+
			"of size %d",
			value, qreg.size)
		panic(err_str)
	}
	for i, _ := range qreg.states {
		qreg.states[i] = complex(0, 0)
	}
	qreg.states[value] = complex(1.0, 0)
}

// Set a particular bit in a QReg
func (qreg *QReg) BSet(index int, value int) {
	if value > 1 {
		err_str := fmt.Sprintf("Value %d should be either 0 or 1",
			value)
		panic(err_str)
	}
	bit := 1 << uint(index)
	bitval := value << uint(index)
	bitnot := (1 - value) << uint(index)
	bprob := qreg.BProb(index, value)
	if bprob > 0 {
		amp_factor := complex(1.0/math.Sqrt(bprob), 0)
		// Alter every state.  If it's the right qubit value, fix the
		// amplitude; otherwise, set the amplitude to 0.
		for state, amp := range qreg.states {
			if int(state)&bit == bitval {
				qreg.states[state] = amp * amp_factor
			} else {
				qreg.states[state] = complex(0, 0)
			}
		}
	} else {
		// Iterate through all the states where this bit is 1
		for state := int(0) | bit; state < int(len(qreg.states)); state = (state + 1) | bit {
			// Add the amplitude of the old state to the new state
			old_state := state - bitval
			new_state := state - bitnot
			qreg.states[new_state] += qreg.states[old_state]
			qreg.states[old_state] = complex(0, 0)
		}
	}
}

// Measure a bit without collapsing its quantum state
func (qreg *QReg) BMeasurePreserve(index int) int {
	if rand.Float64() < qreg.BProb(index, 0) {
		return 0
	}
	return 1
}

// Measure a bit (the quantum state of this qubit will collapse)
func (qreg *QReg) BMeasure(index int) int {
	b := qreg.BMeasurePreserve(index)
	qreg.BSet(index, b)
	return b
}

// Measure a register without collapsing its quantum state
func (qreg *QReg) MeasurePreserve() int {
	r := rand.Float64()
	sum := float64(0.0)
	for i, _ := range qreg.states {
		sum += qreg.StateProb(i)
		if r < sum {
			return i
		}
	}
	return len(qreg.states) - 1
}

// Measure a register
func (qreg *QReg) Measure() int {
	value := qreg.MeasurePreserve()
	var amp complex128
	if real(qreg.states[value]) > 0 {
		amp = complex(1, 0)
	} else {
		amp = complex(-1, 0)
	}
	for i, _ := range qreg.states {
		qreg.states[i] = complex(0, 0)
	}
	qreg.states[value] = amp
	return value
}

func (qreg *QReg) PrintState(index int) {
	prob := qreg.StateProb(index)
	largest := (1 << uint(qreg.size)) - 1
	padding := int(math.Floor(math.Log10(float64(largest)))) + 1
	format := fmt.Sprintf("%%+f%%f|(%%%dd)%%0%db>", padding, qreg.size)
	fmt.Printf(format, qreg.states[index], prob, index, index)
}

func (qreg *QReg) PrintStateln(index int) {
	qreg.PrintState(index)
	fmt.Println()
}

func (qreg *QReg) Print() {
	for i, _ := range qreg.states {
		qreg.PrintStateln(i)
	}
}

func (qreg *QReg) PrintNonZero() {
	for i, state := range qreg.states {
		if state != 0 {
			qreg.PrintStateln(i)
		}
	}
}
