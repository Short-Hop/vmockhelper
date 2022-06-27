# VRender

Originally a fork of [go-render](https://github.com/luci/go-render/issues) 

Implements a more verbose form of the standard Go string formatter, `fmt.Sprintf("%#v", value)`, adding:
- Pointer recursion. Normally, Go stops at the first pointer and prints its
  address. The *render* package will recurse and continue to render pointer
  values.
- Custom type name rendering.
- Deterministic key sorting for `string`- and `int`-keyed maps.

Additional changes made to original fork for include:
- Printing `Time` objects as `time.Date()` equivalent calls
- Ignoring unexported fields that cannot be set as literals
- Printing byte arrays as their string equivalent

Ideally, calling `vrender.Render()` on any object in memory should print a golang code representation of that data