package BitFlip

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"

	mathEth "github.com/ethereum/go-ethereum/common/math"
)

type IBitFlip interface {
	Initalize(pstrTestType string, pITestCount interface{}, parrErrRates []float64, pOutput Output)
	BitFlip(pbigNum *big.Int, pdecRate float64, plngFlipCount int) (*big.Int, Iteration)
}

type ErrorData struct {
	PreviousValue *big.Int
	PreviousByte  string
	IntBits       []int
	ErrorValue    *big.Int
	ErrorByte     string
	DeltaValue    *big.Int
	When          string
}

type Iteration struct {
	IterationNum int
	ErrorData    ErrorData
}

type ErrorRate struct {
	Rate     float64
	FlipData []Iteration
}

type Output struct {
	Data []ErrorRate
}

var (
	mstrTestType    string
	mlngIterations  int
	mlngVarsChanged int
	mdurNanoSeconds time.Duration
	mtimStartTime   time.Time
	marrErrRates    []float64
	mintRateIndex   int = 0
)

// Set up the testing environment with the test type, number of
// changes/iterations or duration in seconds, and error rates. This is
// PER error rate. i.e. 5 minutes and ten error rates will be 50 minutes.
// Test types:
// 'iteration' - increments for each bit flipped
// 'variable' - increments for each variable, regardless of bits flipped
// 'time' - checks against passage of time since started
func Initalize(pstrTestType string, pITestCount interface{}, parrErrRates []float64, pOutput Output) {
	mstrTestType = pstrTestType
	switch mstrTestType {
	case "iteration":
		mlngIterations = pITestCount.(int)
	case "variable":
		mlngVarsChanged = pITestCount.(int)
	default:
		mtimStartTime = time.Now()
		mdurNanoSeconds = time.Duration(pITestCount.(float64) * math.Pow(10, 9))
	}
	var flipData []Iteration
	for _, errRate := range parrErrRates {
		Rate := ErrorRate{errRate, flipData}
		pOutput.Data = append(pOutput.Data, Rate)
	}
}

// BitFlip will run the odds of flipping a bit within pbigNum based on error
// rate pdecRate. The iteration count will increment and both the new number
// and the iteration error data will be returned.
func (this *Output) BitFlip(pbigNum *big.Int, plngFlipCount int) *big.Int {
	rand.Seed(time.Now().UnixNano())

	// Check for out of bounds
	if plngFlipCount > mlngIterations ||
		plngFlipCount > mlngVarsChanged ||
		time.Since(mtimStartTime) >= mdurNanoSeconds {
		if mintRateIndex < len(marrErrRates) {
			mintRateIndex++
		}
		return pbigNum
	}

	decRate := (*this).Data[mintRateIndex].Rate
	var arrBits []int

	// Store previous states
	bigPrevNum, _ := new(big.Int).SetString(pbigNum.String(), 10)
	bigPrevNum = mathEth.U256(bigPrevNum)
	bytPrevNum := bigPrevNum.Bytes()
	bytNum := pbigNum.Bytes()

	// Run chance of flipping a bit in byte representation
	for i, byt := range bytNum {
		for j := 0; j < 8; j++ {
			if math.Floor(rand.Float64()/decRate) == math.Floor(rand.Float64()/decRate) {
				plngFlipCount++
				arrBits = append(arrBits, (i*8)+j)
				bytNum[i] = byt ^ (1 << j)
			}
		}
	}

	// Recreate number from byte code
	pbigNum.SetBytes(bytNum)

	// Build error data
	iteration := Iteration{
		int(plngFlipCount),
		ErrorData{
			bigPrevNum,
			hex.EncodeToString(bytPrevNum),
			arrBits,
			pbigNum,
			hex.EncodeToString(bytNum),
			big.NewInt(0).Sub(pbigNum, bigPrevNum),
			time.Now().Format("01-02-2006-15:04:05.000000000"),
		},
	}

	// Pretty print JSON in console and append to error rate data
	bytJSON, _ := json.MarshalIndent(iteration, "", "    ")
	fmt.Println(string(bytJSON))

	(*this).Data[mintRateIndex].FlipData = append((*this).Data[mintRateIndex].FlipData, iteration)

	return pbigNum
}

func (this Output) MarshalIndent() string {
	byt, err := json.MarshalIndent(this, "", "\t")
	if err != nil {
		return string(byt)
	}
	return "err"
}
