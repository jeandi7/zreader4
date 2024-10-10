package zinterpreter

// zanzibar restricted BNF grammar

/**
<Zschema> ::= <Zdef>*
<Zdef> ::= "definition" <Zname> "{" <Zbody> "}"  ---> generation
<Zname> ::= <identifier>
<Zbody> ::= <Zrelation>*
<Zrelation> ::= "relation" <Rname> ":" <Sname> ("|" <Sname)*   ---> generation
<Rname> ::= <identifier>
<Sname> ::= <Zname> | <Zname> "#" <Rname> | <Zname> ":" "*"
<identifier> ::= [a-zA-Z_][a-zA-Z0-9_]*

*/

import (
	"fmt"
	"strings"
	"unicode"
)

// Token represents the different tokens
type Token int

const (
	DefinitionToken Token = iota // "definition"
	RelationToken                // "relation"
	ColonToken                   // ":"
	OrToken                      // "|"
	LeftBraceToken               // "{"
	RightBraceToken              // "}"
	HashToken                    // "#"
	IdentifierToken              // [a-zA-Z_][a-zA-Z0-9_]*
	WildCard                     // *
	EOFToken                     // ''
	InvalidToken                 //
)

// Item représente un token avec sa valeur
type Item struct {
	Token Token
	Value string
}

// Lexer parses input text and generates tokens
type Lexer struct {
	input       string
	pos         int
	length      int
	currentItem *Item
}

// for Lexer message
func TokenToString(t Token) string {
	switch t {
	case DefinitionToken:
		return "definition"
	case RelationToken:
		return "relation"
	case ColonToken:
		return ":"
	case OrToken:
		return "|"
	case LeftBraceToken:
		return "{"
	case RightBraceToken:
		return "}"
	case HashToken:
		return "#"
	case IdentifierToken:
		return "Identifier"
	case WildCard:
		return "*"
	case EOFToken:
		return ""
	case InvalidToken:
		return "invalid"
	default:
		return "unknown"
	}
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		length: len(input),
		currentItem: &Item{
			Token: InvalidToken,
			Value: "",
		},
	}
}

// We eat up the white spaces
func (l *Lexer) eatSpace() {
	for l.pos < l.length && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
}

// Lexer returns the next token to read
func (l *Lexer) NextToken() *Item {
	l.eatSpace()

	if l.pos >= l.length {
		l.currentItem.Token = EOFToken
		l.currentItem.Value = ""
		return l.currentItem
	}

	switch {
	case strings.HasPrefix(l.input[l.pos:], "definition"):
		l.currentItem.Token = DefinitionToken
		l.currentItem.Value = "definition"
		l.pos += len("definition")
	case strings.HasPrefix(l.input[l.pos:], "relation"):
		l.currentItem.Token = RelationToken
		l.currentItem.Value = "relation"
		l.pos += len("relation")
	case l.input[l.pos] == ':':
		l.currentItem.Token = ColonToken
		l.currentItem.Value = ":"
		l.pos++
	case l.input[l.pos] == '|':
		l.currentItem.Token = OrToken
		l.currentItem.Value = "|"
		l.pos++
	case l.input[l.pos] == '{':
		l.currentItem.Token = LeftBraceToken
		l.currentItem.Value = "{"
		l.pos++
	case l.input[l.pos] == '}':
		l.currentItem.Token = RightBraceToken
		l.currentItem.Value = "}"
		l.pos++
	case l.input[l.pos] == '#':
		l.currentItem.Token = HashToken
		l.currentItem.Value = "#"
		l.pos++
	case l.input[l.pos] == '*':
		l.currentItem.Token = WildCard
		l.currentItem.Value = "*"
		l.pos++

	default:
		if unicode.IsLetter(rune(l.input[l.pos])) {
			start := l.pos
			for l.pos < l.length && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
				l.pos++
			}
			l.currentItem.Token = IdentifierToken
			l.currentItem.Value = l.input[start:l.pos]
		} else {
			l.currentItem.Token = InvalidToken
			l.currentItem.Value = string(l.input[l.pos])
			l.pos++
		}
	}
	return l.currentItem
}

func (l *Lexer) readAndMatchToken(expected Token) error {
	if l.currentItem.Token == expected {
		return nil
	}
	return fmt.Errorf("expected token '%v', but got '%v'", TokenToString(expected), l.currentItem.Value)
}

// Syntaxic Analyser
type ZDef struct {
	Name      string
	Relations []*ZRelation
	ID        string
}

type ZRelation struct {
	Name             string
	Zobjects         []*Zobject
	ZobjectSets      []*ZobjectSet
	ZobjectWildCards []*ZobjectWildCard
	ID               string
	myZDef           *ZDef
}

// object
type Zobject struct {
	Name   string
	ID     string
	myZDef *ZDef
	Unique bool
}

// object#relation
type ZobjectSet struct {
	Name       string
	Relation   string
	ID         string
	IDRelation string
	Unique     bool
}

// object:*
type ZobjectWildCard struct {
	Name   string
	ID     string
	Unique bool
}

// object#relation

// <Zschema> ::= <Zdef>*
func (l *Lexer) ReadZSchema() ([]*ZDef, error) {
	var zdefs []*ZDef

	for l.currentItem.Token != EOFToken {
		_zdef, _err := l.readZDef()
		if _err != nil {
			return zdefs, _err
		}
		zdefs = append(zdefs, &_zdef)
		l.NextToken()

	}
	return zdefs, nil
}

// <Zdef> ::= "definition" <Zname> "{" <Zbody> "}"
func (l *Lexer) readZDef() (ZDef, error) {
	var zdef ZDef

	// read "definition"
	err := l.readAndMatchToken(DefinitionToken)
	if err != nil {
		return zdef, err
	}
	l.NextToken()

	// read <Zname>
	err = l.readAndMatchToken(IdentifierToken)
	if err != nil {
		return zdef, err
	}
	zdef.Name = l.currentItem.Value
	l.NextToken()

	// read '{'
	err = l.readAndMatchToken(LeftBraceToken)
	if err != nil {
		return zdef, err
	}
	l.NextToken()

	// read ZBody
	// ZBody is not a token
	// no need to call NextToken after

	zdef, err = l.readZBody(zdef)
	if err != nil {
		return zdef, err
	}

	// read '}'
	err = l.readAndMatchToken(RightBraceToken)
	if err != nil {
		return zdef, err
	}

	return zdef, nil
}

// <Zbody> ::= <Zrelation>*
// * means zero or more <Zrelation>
func (l *Lexer) readZBody(zdef ZDef) (ZDef, error) {
	// var zdef ZDef

	for l.currentItem.Token == RelationToken {

		relation, err := l.readZRelation()
		if err != nil {
			return zdef, err
		}
		zdef.Relations = append(zdef.Relations, &relation)
	}

	return zdef, nil
}

// <Zrelation> ::= "relation" <Rname> ":" <Sname> ("|" <Sname)*
// <Sname> ::= <Zname> | <Zname> "#" <Rname> | <Zname> ":" "*"

func (l *Lexer) readZRelation() (ZRelation, error) {
	var zrelation ZRelation

	if l.currentItem.Value != "relation" {
		return zrelation, fmt.Errorf("expected 'relation', but got '%s'", l.currentItem.Value)
	}
	l.NextToken()

	err := l.readAndMatchToken(IdentifierToken)
	if err != nil {
		return zrelation, err
	}
	zrelation.Name = l.currentItem.Value
	l.NextToken()

	err = l.readAndMatchToken(ColonToken)
	if err != nil {
		return zrelation, err
	}
	l.NextToken()

	for l.currentItem.Token == IdentifierToken {
		_name := l.currentItem.Value
		l.NextToken()

		//  <Sname> ::= <Zname> "#" <Rname>
		if l.currentItem.Token == HashToken {
			l.NextToken()
			err := l.readAndMatchToken(IdentifierToken) // <Rname>
			if err != nil {
				return zrelation, err
			} else {
				zrelation.ZobjectSets = append(zrelation.ZobjectSets, &ZobjectSet{Name: _name, Relation: l.currentItem.Value})
				l.NextToken()
			}

		} else {
			// <Sname> ::= <Zname> ":" "*"
			if l.currentItem.Token == ColonToken {
				l.NextToken()
				err := l.readAndMatchToken(WildCard) // "*"
				if err != nil {
					return zrelation, err
				} else {
					// zrelation.ZobjectAll =
					zrelation.ZobjectWildCards = append(zrelation.ZobjectWildCards, &ZobjectWildCard{Name: _name})
					l.NextToken()
				}
			} else { // <Sname> ::= <Zname>
				zrelation.Zobjects = append(zrelation.Zobjects, &Zobject{Name: _name})
			}
		}

		if l.currentItem.Token == OrToken {
			l.NextToken()
		} else {
			break
			// it's the last
		}
	}

	return zrelation, nil
}

// Generation Code

type PlantUMLArchimateSchema struct {
	Zdefs   []*ZDef
	ZdefMap map[string]*ZDef

	SchemaDpi   int
	SchemaScale float64
}

// Generate a row for each businessObject
/*
func (plantUMLArchimateSchema *PlantUMLArchimateSchema) generateRowForEachBusinessObject(out []string) {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		line := fmt.Sprintf("Business_Object(%s,\"%s\")", zdef.ID, zdef.Name)
		out = append(&out, line)
	}
}
*/

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) Generate(pngfilename string) string {
	var out []string
	plantUMLArchimateSchema.createIDforZdef()

	out = append(out, "@startuml "+pngfilename)
	out = append(out, "!include <archimate/Archimate>")

	out = append(out, "scale 1.0")
	out = append(out, "skinparam dpi 96")

	// Generate a row for each businessObject

	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		line := fmt.Sprintf("Business_Object(%s,\"%s\")", zdef.ID, zdef.Name)
		out = append(out, line)
	}

	// Generate a relationship line as a business object for each zdef
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			switch zrel.ID {
			case "NOTDRAW":
				line := fmt.Sprintf("rectangle \"relation %s is duplicated in definition %s \" #red", zrel.Name, zdef.Name)
				out = append(out, line)
			default:
				line := fmt.Sprintf("Business_Object(%s,\"%s\") <<relation>>", zrel.ID, zrel.Name)
				line2 := fmt.Sprintf("Rel_Association(%s,%s)", zdef.ID, zrel.ID)
				out = append(out, line)
				out = append(out, line2)
				for _, zobject := range zrel.Zobjects {
					switch zobject.ID {
					case "NOTDRAW":
						line3 := fmt.Sprintf("rectangle \"definition %s does not exist \" #red", zobject.Name)
						out = append(out, line3)
					default:
						switch zobject.Unique {
						case true:
							line4 := fmt.Sprintf("Rel_Access_w(%s,%s)", zrel.ID, zobject.ID)
							out = append(out, line4)
						default:
							line4 := fmt.Sprintf("rectangle \" %s is declared more that one in relation %s of definition %s\" #red ", zdef.Name, zrel.Name, zdef.Name)
							out = append(out, line4)
						}
					}
				}
			}
		}
	}

	// Generate a relationshipSet row on a relation
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			for _, zobjectSet := range zrel.ZobjectSets {
				switch zobjectSet.ID {
				case "NOTDRAW":
					line := fmt.Sprintf("rectangle \"definition %s does not exist in \" #red", zobjectSet.Name)
					out = append(out, line)
				default:
					switch zobjectSet.IDRelation {
					case "NOTDRAW":
						line := fmt.Sprintf("rectangle \"  %s#%s in definition %s  : relation %s does not exist in %s \"  #red", zobjectSet.Name, zobjectSet.Relation, zdef.Name, zobjectSet.Relation, zobjectSet.Name)
						out = append(out, line)
					default:
						switch zobjectSet.Unique {
						case true:
							line2 := fmt.Sprintf("Rel_Access_w(%s,%s,\"%s#%s\")", zobjectSet.IDRelation, zrel.ID, zobjectSet.Name, zobjectSet.Relation)
							out = append(out, line2)
						case false:
							line2 := fmt.Sprintf("rectangle \"  %s#%s declared more that one in relation %s of definition %s \"  #red", zobjectSet.Name, zobjectSet.Relation, zrel.Name, zdef.Name)
							out = append(out, line2)
						}
					}
				}
			}
		}
	}

	// Generate a relationWildCard row on a relation
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			for _, zobjectWildCard := range zrel.ZobjectWildCards {
				switch zobjectWildCard.ID {
				case "NOTDRAW":
					line := fmt.Sprintf("rectangle \"definition %s does not exist in \" #red", zobjectWildCard.Name)
					out = append(out, line)
				default:
					switch zobjectWildCard.Unique {

					case true:
						line2 := fmt.Sprintf("Rel_Access_w(%s,%s,\"%s\")", zrel.ID, zobjectWildCard.ID, "ALL")
						out = append(out, line2)

					case false:
						line3 := fmt.Sprintf("rectangle \"wildcard  %s:* is declared more than one in relation %s of definition %s\" #red", zobjectWildCard.Name, zrel.Name, zdef.Name)
						out = append(out, line3)
					}

				}
			}
		}
	}

	out = append(out, "@enduml")
	return strings.Join(out, "\n")
}

// utility
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) createIDforZdef() {
	zdefMapNameToVarName := make(map[string]string)

	for index, zdef := range plantUMLArchimateSchema.Zdefs {
		varname := fmt.Sprintf("b%d", index+1)
		if _, exists := zdefMapNameToVarName[zdef.Name]; exists {
			fmt.Printf("definition %s is declared more that one  \n", zdef.Name)
			continue
		}
		zdefMapNameToVarName[zdef.Name] = varname
		zdef.ID = varname
	}

	plantUMLArchimateSchema.createIDforZdefRelations()
	plantUMLArchimateSchema.initZdefMap()
	plantUMLArchimateSchema.verifyAndAssignIDforZobjectInRelations()
	plantUMLArchimateSchema.verifyUniqueObjectForEachRelation()
	plantUMLArchimateSchema.verifyAndAssignIDInRelationsforZobjectSet()
	plantUMLArchimateSchema.verifyUniqueSetObjectForEachRelation()
	plantUMLArchimateSchema.verifyAndAssignIDInRelationsforZobjectWildCard()
	plantUMLArchimateSchema.verifyUniqueObjectWildCardForEachRelation()
}

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) createIDforZdefRelations() {
	var relCount int = 0
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		RelNameSlice := []string{}
		for _, zrel := range zdef.Relations {
			relCount++
			varname := fmt.Sprintf("r%d", relCount)
			if contains(RelNameSlice, zrel.Name) {
				zrel.ID = "NOTDRAW"
				fmt.Printf("relation %s is declared more that one in definition %s \n", zrel.Name, zdef.Name)

			} else {
				RelNameSlice = append(RelNameSlice, zrel.Name)
				zrel.ID = varname
				zrel.myZDef = zdef
			}
		}
	}

}

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) initZdefMap() {
	plantUMLArchimateSchema.ZdefMap = make(map[string]*ZDef)
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		plantUMLArchimateSchema.ZdefMap[zdef.Name] = zdef
	}

}

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) verifyAndAssignIDforZobjectInRelations() {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			for _, zobject := range zrel.Zobjects {
				if myZDef, exists := plantUMLArchimateSchema.ZdefMap[zobject.Name]; exists {
					zobject.ID = myZDef.ID
					zobject.myZDef = myZDef
				} else {
					fmt.Printf("%s declared in relation %s of definition %s does not exist.  \n", zobject.Name, zrel.Name, zdef.Name)
					zobject.ID = "NOTDRAW"
				}

			}
		}
	}
}

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) findZDef(objectName string) (*ZDef, error) {
	if myZDef, exists := plantUMLArchimateSchema.ZdefMap[objectName]; exists {
		return myZDef, nil
	} else {
		err := fmt.Errorf("definition %s does not exist. ", objectName)
		return nil, err
	}

}

// with tuple example like object:id#relation1@objectSet#relation2 (with Zanzibar notation)
// objectName is objectSet, relationName is relation2

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) findZRelation(objectName string, relationName string) (*ZRelation, error) {
	myZf, err := plantUMLArchimateSchema.findZDef(objectName)
	if err != nil {
		return nil, err
	} else {
		zrelMap := make(map[string]*ZRelation)
		for _, zrel := range myZf.Relations {
			zrelMap[zrel.Name] = zrel
		}
		if myZrel, exists := zrelMap[relationName]; exists {
			return myZrel, nil
		} else {
			err = fmt.Errorf("relation %s does not exist in definition %s. ", relationName, objectName)
			return nil, err
		}
	}

}

// tuple like resource:id#relation@group#relation (with Zanzibar notation)
// group#relation must exist

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) verifyAndAssignIDInRelationsforZobjectSet() {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			for _, zobjectSet := range zrel.ZobjectSets {
				if myZDef, exists := plantUMLArchimateSchema.ZdefMap[zobjectSet.Name]; exists {
					zobjectSet.ID = myZDef.ID
					myZel, error := plantUMLArchimateSchema.findZRelation(myZDef.Name, zobjectSet.Relation)
					if error != nil {
						zobjectSet.IDRelation = "NOTDRAW"
						fmt.Printf("relation %s declared in %s does not exist in  %s.  \n", zobjectSet.Relation, zdef.Name, myZDef.Name)

					} else {
						//double ?
						zobjectSet.IDRelation = myZel.ID
					}

				} else {
					fmt.Printf("%s declared in %s does not exist.  \n", zobjectSet.Name, zdef.Name)
					zobjectSet.ID = "NOTDRAW"
				}

			}
		}
	}
}

// tuple like resource:id#relation@user:* (with Zanzibar notation)

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) verifyAndAssignIDInRelationsforZobjectWildCard() {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			for _, zobjectWildCard := range zrel.ZobjectWildCards {
				if myZDef, exists := plantUMLArchimateSchema.ZdefMap[zobjectWildCard.Name]; exists {
					zobjectWildCard.ID = myZDef.ID

				} else {
					fmt.Printf("%s declared in relation %s of definition %s does not exist.  \n", zobjectWildCard.Name, zrel.Name, zdef.Name)
					zobjectWildCard.ID = "NOTDRAW"
				}
			}
		}
	}
}

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) verifyUniqueObjectForEachRelation() {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			keyObjectSlice := []string{}
			for _, zobject := range zrel.Zobjects {
				varname := zobject.Name
				if contains(keyObjectSlice, varname) {
					zobject.Unique = false
					fmt.Printf("%s is declared more that one in relation %s of definition %s\n", varname, zrel.Name, zdef.Name)
				} else {
					zobject.Unique = true
					keyObjectSlice = append(keyObjectSlice, varname)
				}
			}
		}
	}
}

// 'objectName#relationName' must be unique in zrel

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) verifyUniqueSetObjectForEachRelation() {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			keySetObjectSlice := []string{}
			for _, zobjectSet := range zrel.ZobjectSets {
				varname := fmt.Sprintf("%s#%s", zobjectSet.Name, zobjectSet.Relation)
				if contains(keySetObjectSlice, varname) {
					zobjectSet.Unique = false
					fmt.Printf("%s is declared more that one in relation %s of definition %s\n", varname, zrel.Name, zdef.Name)
				} else {
					zobjectSet.Unique = true
					keySetObjectSlice = append(keySetObjectSlice, varname)
				}
			}
		}
	}
}

// 'objectname:*' must be unique in zrel

func (plantUMLArchimateSchema *PlantUMLArchimateSchema) verifyUniqueObjectWildCardForEachRelation() {
	for _, zdef := range plantUMLArchimateSchema.Zdefs {
		for _, zrel := range zdef.Relations {
			keySetObjectSlice := []string{}
			for _, zobjectWildCard := range zrel.ZobjectWildCards {
				varname := zobjectWildCard.Name
				if contains(keySetObjectSlice, varname) {
					zobjectWildCard.Unique = false
					fmt.Printf("wildcard %s:* is declared more that one in relation %s of definition %s\n", varname, zrel.Name, zdef.Name)
				} else {
					zobjectWildCard.Unique = true
					keySetObjectSlice = append(keySetObjectSlice, varname)
				}
			}
		}
	}
}
