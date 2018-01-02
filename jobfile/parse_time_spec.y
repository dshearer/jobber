%{

package jobfile

import (
    "fmt"
    "github.com/dshearer/jobber/common"
    "sort"
    "strconv"
    "strings"
    "unicode"
)

const (
    gEof = 0
)

type TimeSpecExp interface{
    Eval(fieldName string, min int, max int) (TimeSpec, error)
}

type OneValTimeSpecExp struct {
    val int
}

func (self *OneValTimeSpecExp) Eval(fieldName string, min int,
    max int) (TimeSpec, error) {

    errMsg := fmt.Sprintf("Invalid '%v' value", fieldName)

    // check range
    if self.val < min {
        errMsg := fmt.Sprintf("%s: cannot be less than %v.", errMsg, min)
        return nil, &common.Error{What: errMsg}
    } else if self.val > max {
        errMsg := fmt.Sprintf("%s: cannot be greater than %v.", errMsg, max)
        return nil, &common.Error{What: errMsg}
    }
    
    return &OneValTimeSpec{val: self.val}, nil
}

type WildcardTimeSpecExp struct {}

func (self *WildcardTimeSpecExp) Eval(fieldName string, min int,
    max int) (TimeSpec, error) {
    
    return &WildcardTimeSpec{}, nil
}

type SetTimeSpecExp struct {
    setExp SetExp
}

func (self *SetTimeSpecExp) Eval(fieldName string, min int,
    max int) (TimeSpec, error) {
    
    vals, err := self.setExp.Eval(fieldName, min, max)
    if err != nil {
        return nil, err
    }
    return &SetTimeSpec{vals: vals, desc: self.setExp.String()}, nil
}

type RandomTimeSpecExp struct {
    setExp SetExp
}

func (self *RandomTimeSpecExp) Eval(fieldName string, min int,
    max int) (TimeSpec, error) {
    
    vals, err := self.setExp.Eval(fieldName, min, max)
    if err != nil {
        return nil, err
    }
    desc := fmt.Sprintf("R%v", self.setExp)
    return &RandomTimeSpec{vals: vals, desc: desc}, nil
}

type SetExp interface {
    fmt.Stringer
    Eval(fieldName string, min, max int) ([]int, error)
}

type StepSetExp struct {
    step int
}

func (self *StepSetExp) String() string {
    return fmt.Sprintf("*/%v", self.step)
}

func (self *StepSetExp) Eval(fieldName string, min int,
    max int) ([]int, error) {
    
    var vals []int
    for v := min; v <= max; v = v + self.step {
        vals = append(vals, v)
    }
    return vals, nil
}

type EnumSetExp struct {
    ints []int
}

func (self *EnumSetExp) normValues() []int {
    sortedInts := append(make([]int, 0, len(self.ints)), self.ints...)
    sort.Ints(sortedInts)
    
    uniqMap := make(map[int]bool)
    var uniqInts []int
    for _, v := range sortedInts {
        if _, ok := uniqMap[v]; !ok {
            uniqInts = append(uniqInts, v)
            uniqMap[v] = true
        }
    }
    return uniqInts
}

func (self *EnumSetExp) String() string {
    var strs []string
    for _, i := range self.normValues() {
        strs = append(strs, fmt.Sprintf("%v", i))
    }
    return strings.Join(strs, ",")
}

func (self *EnumSetExp) Eval(fieldName string, min int,
    max int) ([]int, error) {
    
    errMsgPrefix := fmt.Sprintf("Invalid \"%v\" value", fieldName)
    
    // check values
    for _, v := range self.ints {
        if v < min {
            msg := fmt.Sprintf("%v: Values must be greater than or " +
                "equal to %v", errMsgPrefix, min)
            return nil, &common.Error{What: msg}
        } else if v > max {
            msg := fmt.Sprintf("%v: Values must be less than or " +
                "equal to %v", errMsgPrefix, max)
            return nil, &common.Error{What: msg}
        }
    }
    
    return self.normValues(), nil
}

type RangeSetExp struct {
    start int
    end   int
}

func (self *RangeSetExp) String() string {
    return fmt.Sprintf("%v-%v", self.start, self.end)
}

func (self *RangeSetExp) Eval(fieldName string, min int,
    max int) ([]int, error) {
    
    errMsgPrefix := fmt.Sprintf("Invalid \"%v\" value", fieldName)
    
    // check values
    if self.start < min {
        msg := fmt.Sprintf("%v: Values must be greater than or " +
            "equal to %v", errMsgPrefix, min)
        return nil, &common.Error{What: msg}
    } else if self.end > max {
        msg := fmt.Sprintf("%v: Values must be less than or " +
            "equal to %v", errMsgPrefix, max)
        return nil, &common.Error{What: msg}
    } else if self.start > self.end {
        msg := fmt.Sprintf("%s: start must be less than or " + 
            "equal to end", errMsgPrefix)
        return nil, &common.Error{What: msg}
    }
    
    // make values
    var vals []int
    for i := self.start; i <= self.end; i++ {
        vals = append(vals, i)
    }
    return vals, nil
}

type AnySetExp struct{}

func (self *AnySetExp) String() string {
    return ""
}

func (self *AnySetExp) Eval(fieldName string, min int,
    max int) ([]int, error) {
    
    var vals []int
    for i := min; i <= max; i++ {
        vals = append(vals, i)
    }
    return vals, nil
}

var gPhrase   TimeSpecExp
var gErrorMsg *string

func setErrorMsg(msg string) {
    if gErrorMsg == nil {
        gErrorMsg = &msg
    } else {
        newMsg := fmt.Sprintf("%v: %v", msg, *gErrorMsg)
        gErrorMsg = &newMsg
    }
}

%}

%union {
    nbr                 *int
    randTimeSpecExp     *RandomTimeSpecExp
    setExp              SetExp
    stepSetExp          *StepSetExp
    enumSetExp          *EnumSetExp
    rangeSetExp         *RangeSetExp
}

%type <randTimeSpecExp>   rand_time_spec_exp
%type <setExp>            set_exp
%type <stepSetExp>        step_set_exp
%type <enumSetExp>        enum_set_exp
%type <enumSetExp>        enum_set_exp_tail
%type <rangeSetExp>       range_set_exp

%token ',' '-' '*' 'R' STAR_SLASH
%token <nbr>   INT

%%

top:
    INT
    { gPhrase = &OneValTimeSpecExp{val: *$1} }
|   set_exp
    { gPhrase = &SetTimeSpecExp{setExp: $1} }
|   '*'
    { gPhrase = &WildcardTimeSpecExp{} }
|   rand_time_spec_exp
    { gPhrase = $1 }

rand_time_spec_exp:
    'R' set_exp
    { $$ = &RandomTimeSpecExp{setExp: $2} }
|   'R'
    { $$ = &RandomTimeSpecExp{setExp: &AnySetExp{}} }

set_exp:
    step_set_exp
    { $$ = $1 }
|   enum_set_exp
    { $$ = $1 }
|   range_set_exp
    { $$ = $1 }

step_set_exp:
    STAR_SLASH INT
    { $$ = &StepSetExp{step: *$2} }
|   STAR_SLASH error
    {
        setErrorMsg("Expected int after \"*/\"")
        goto ret1
    }

enum_set_exp:
    INT enum_set_exp_tail
    {
        $$ = &EnumSetExp{ints: []int{*$1}}
        $$.ints = append($$.ints, $2.ints...)
    }

enum_set_exp_tail:
    ',' INT
    { $$ = &EnumSetExp{ints: []int{*$2}} }
|   ',' INT enum_set_exp_tail
    {
        $$ = &EnumSetExp{ints: []int{*$2}}
        $$.ints = append($$.ints, $3.ints...)
    }
|   ',' error
    {
        setErrorMsg("Expected int after \",\"")
        goto ret1
    }

range_set_exp:
    INT '-' INT
    { $$ = &RangeSetExp{start: *$1, end: *$3} }
|   INT '-' error
    {
        setErrorMsg("Expected int after \"-\"")
        goto ret1
    }
%%

type yyLex struct {
    expr      string
    peek      rune
}

func NewTimeSpecLexer(expr string) (*yyLex) {
    return &yyLex{expr: expr}
}

func (self *yyLex) Error(msg string) {
    if gErrorMsg == nil {
        setErrorMsg(msg)
    }
}

func (self *yyLex) Lex(yylval *yySymType) int {
    for {
        r := self.nextRune()
        switch r {
        case ',', '-', 'R':
            return int(r)
            
        case '*':
            r2 := self.nextRune()
            if r2 == '/' {
                return STAR_SLASH
            } else {
                self.peek = r2
                return int('*')
            }
            
        case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
            numeral := string(r)
        ReadNums:
            for {
                r2 := self.nextRune()
                if unicode.IsDigit(r2) {
                    numeral += string(r2)
                } else if unicode.IsSpace(r2) || unicode.IsPunct(r2) || r2 == gEof {
                    self.peek = r2
                    break ReadNums
                } else {
                    numeral += string(r2)
                    setErrorMsg(fmt.Sprintf("Invalid number: \"%v\"", numeral))
                    return gEof
                }
            }
            number, err := strconv.Atoi(numeral)
            if err != nil {
                setErrorMsg(fmt.Sprintf("Invalid number: \"%v\"", numeral))
                return gEof
            }
            yylval.nbr = &number
            return INT
            
        case ' ', '\t':
            // ignore
        
        case gEof:
            return gEof
            
        default:
            var msg string
            if unicode.IsGraphic(r) {
                msg = fmt.Sprintf("Unexpected char: \"%v\"", string(r))
            } else {
                msg = fmt.Sprintf("Unexpected char: %v", r)
            }
            setErrorMsg(msg)
            return gEof
        }
    }
}

func (self *yyLex) nextRune() rune {
    if self.peek != gEof {
        r := self.peek
        self.peek = gEof
        return r
    } else if len(self.expr) == 0 {
        return gEof
    } else {
        r := self.expr[0]
        self.expr = self.expr[1:]
        return rune(r)
    }
}

func parseTimeSpec(s string, fieldName string, min int,
    max int) (TimeSpec, error) {
    
    // parse
    gErrorMsg = nil
    lex := NewTimeSpecLexer(s)
    retval := yyParse(lex)
    if retval != 0 {
        msg := fmt.Sprintf("Cannot parse time spec for \"%v\": %v",
            fieldName, *gErrorMsg)
        return nil, &common.Error{What: msg}
    }
    
    // eval
    return gPhrase.Eval(fieldName, min, max)
}