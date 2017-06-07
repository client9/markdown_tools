
## normal link

```
[link test](https://www.client9.com/)
```

[link test](https://www.client9.com/)

## can link text be split?

yes.

```
[link 
  test](https://www.client9.com)
```

[link 
  test](https://www.client9.com)

## can link text and link url be on separate lines

No.

```
[link test]
   (https://www.client9.com)
```

[link test]
   (https://www.client9.com)

## link with title

```
[link test](https://www.client9.com/ "client9")
```

[link test](https://www.client9.com/ "client9")

## link with title split on two lines

[link test](https://www.client9.com/ "line 1
   line 2")

```
[link test](https://www.client9.com/ "line 1
   line 2")
```

## link title on separate lint

[link test](https://www.client9.com/ 
 "line 1 line 2")

```
[link test](https://www.client9.com/
 "line 1 line 2")
```

## link title  with parans on new line

[link test](https://www.client9.com/ 
 "line 1 line 2"
)

```
[link test](https://www.client9.com/
 "line 1 line 2"
)
```

## link text/url spacing

[link test] (https://www.client9.com/)

```
[link test] (https://www.client9.com/)
```

## bracket autolink

```
<https://www.client9.com>
```

<https://www.client9.com>

### autolink with parens

```
<https://www.client9.com/?(foo)>
```

<https://www.client9.com/?(foo)>

## naked link

```
https://www.client9.com/?(foo)
```

https://www.client9.com/?(foo)
