{ A few notes on this example:

* `setA` is a procedure call used as a statement.
* `addtwice(...)` is a function call used as an operand.
* The function returns by assigning to its own name.
* Inside bodies, the same old forms are still used:

  * `a := 5;`
  * `a := b;`
  * `a := b + a;`
  * `a := b + 5;`
}


program test;

var
  a, b, c: integer;

procedure setA(x: integer);
var
  localvar: integer;
begin
  localvar := x + 1;
  a := localvar;
end setA;

function addtwice(x, y: integer; z: integer): integer;
var
  t: integer;
begin
  t := x + y;
  addtwice := t + z;
end addtwice;

begin
  a := 5;
  b := 7;
  setA(a);
  c := addtwice(a, b, 3);
  a := c - 4;
end.
