### Tricky things to parse in the Puppet Language

#### Heredoc
Heredoc is a construct that starts on one line, doesn't consume the whole line,
and then has a body that starts on the next line (or after the next heredoc), and
finally ends with an end tag.

###### Example heredoc:

```puppet
['first', @(SECOND), 'third',
  This is the text of the
  second entry
  |-SECOND
  'fourth']
```
This yields and array that is equal to:
```puppet
['first', "This is the text of the\nsecond entry", 'third', 'fourth']
```

When the lexer encounters a heredoc tag '@(SECOND)', it must remember the position after the tag, produce the
text, and finally remember that when lexing resumes at the ',' (after the tag) and eventually hits
the end of line, the next line is actually after then end tag.

To make things even more challenging, a line may contain several heredoc tags.

###### Example heredoc with several entries:

```puppet
['first', @(SECOND), 'third', @(FOURTH), 'fifth',
  This is the text of the
  second entry
  |-SECOND
  And here is the text of the
  fourth entry
  |-FOURTH
  'sixth']
```
This yields and array that is equal to:
```puppet
['first', "This is the text of the\nsecond entry", 'third',
  "And here is the text of the\nfourth entry", 'fifth', 'sixth']
```

#### Interpolations
A double quoted string may contain interpolations in the form "${var}" or simply "$var". When
using the brace delimited form, the contained expression can be an expression. That expression
in turn, may contain braces. It may even contain a string which in turn contains a nested
interpolation:

###### Example of a valid nested interpolation
```puppet
$t = 'the'
$r = 'revealed'
$map = {'ipl' => 'meaning', 42.0 => 'life'}
notice "$t ${map['ipl']} of ${map[42.0]}${[3, " is not ${r}"][1]} here"
```
This code will notice the text "the meaning of life is not revealed here".

Also note that the actual interpolation expressions starts with 'map' and not '$map'. Both will work, but
the lexer must take into account that an expression that starts with an identifier is in
fact a variable when encountered in an interpolation expression.

#### Array or keys to an access expression?

Newlines are sometimes significant in Puppet, but not always.

```puppet
expr['a']

```
is an access expression but:
```puppet
expr
['a']

```
is `expr` followed by an array.

#### Unparameterized argument lists and "statement calls"

In puppet, a call to a function may or may not use parentheses to delimit the arguments. This
poses several challenges. One is that it might a conflict with a resource expression. Consider:
```
somefunc { 'a', 'b' }
```
Is this a resource expression without a title, or is it a function call with a hash parameter? It
is very desirable to know the difference so that an sensible error can be printed in case of the
former. The parser uses a "best effort" here, since it really cannot. Instead a set of well known
"statement calls" are ensured to always result in a call.

Consider:
```
notice
'the argument'
```
or, which is exactly the same thing:
```
notice 'the argument'
```
Is this a block that calls notice without arguments and then returns the string 'the argument'? Nope.
Since "notice" happens to be one of the well known "statement calls", it's a construct that returns the value
of the expression `notice('the argument')`. A pair of parentheses must be inserted after 'notice' to make
it an actual call with no parameters.

The exact same syntax, but with a function name that is not known to be a predefined statement call
will yield a syntax error. Hence, this is illegal:
```puppet
function foo($x) {
  notice($x)
}

foo 'the argument'
```
