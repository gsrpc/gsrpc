package gen4go

var tpl4go = `

{{define "enum"}} {{$Enum := title .Name}}

//{{$Enum}} type define -- generate by gsc
type {{$Enum}} {{enumType .}}

//enum {{$Enum}} constants -- generate by gsc
const (
    {{range .Constants}}
    {{$Enum}}{{title .Name}} {{$Enum}} = {{.Value}}
    {{end}}
)

//Write{{$Enum}} write enum to output stream
func Write{{$Enum}}(writer gorpc.Writer, val {{$Enum}}) error{
    return {{if enumSize . | eq 4}} gorpc.WriteUint32(writer,uint32(val)) {{else}} gorpc.WriteByte(writer,byte(val)) {{end}}
}

//Read{{$Enum}} write enum to output stream
func Read{{$Enum}}(reader gorpc.Reader)({{$Enum}}, error){
    val,err := {{if enumSize . | eq 4}} gorpc.ReadUint32(reader) {{else}} gorpc.ReadByte(reader) {{end}}
    return {{$Enum}}(val),err
}

//String implement Stringer interface
func (val {{$Enum}}) String() string {
    switch val {
        {{range .Constants}}
        case {{.Value}}:
            return "enum({{$Enum}}.{{title .Name}})"
        {{end}}
    }
    return fmt.Sprintf("enum(Unknown(%d))",val)
}

{{end}}

{{define "exception"}} {{$Table := title .Name}}

//{{$Table}} -- generate by gsc
type {{$Table}} struct {
    {{range .Fields}}
    {{title .Name}} {{typeName .Type}}
    {{end}}
}

//Error implement error interface
func (e * {{$Table}}) Error() string {
    return "{{$Table}} error"
}

//New{{$Table}} create new struct object with default field val -- generate by gsc
func New{{$Table}}() *{{$Table}} {
    return &{{$Table}}{
        {{range .Fields}}
        {{title .Name}}: {{defaultVal .Type}},
        {{end}}
    }
}

//Read{{$Table}} read {{$Table}} from input stream -- generate by gsc
func Read{{$Table}}(reader gorpc.Reader) (target *{{$Table}},err error) {
    target = New{{$Table}}()
    {{range .Fields}}
    target.{{title .Name}},err = {{readType .Type}}(reader)
    if err != nil {
        return
    }
    {{end}}
    return
}


//Write{{$Table}} write {{$Table}} to output stream -- generate by gsc
func Write{{$Table}}(writer gorpc.Writer,val *{{$Table}}) (err error) {
    {{range .Fields}}
    err = {{writeType .Type}}(writer,val.{{title .Name}})
    if err != nil {
        return
    }
    {{end}}
    return nil
}

{{end}}

{{define "table"}} {{$Table := title .Name}}

//{{$Table}} -- generate by gsc
type {{$Table}} struct {
    {{range .Fields}}
    {{title .Name}} {{typeName .Type}}
    {{end}}
}

//New{{$Table}} create new struct object with default field val -- generate by gsc
func New{{$Table}}() *{{$Table}} {
    return &{{$Table}}{
        {{range .Fields}}
        {{title .Name}}: {{defaultVal .Type}},
        {{end}}
    }
}

//Read{{$Table}} read {{$Table}} from input stream -- generate by gsc
func Read{{$Table}}(reader gorpc.Reader) (target *{{$Table}},err error) {
    target = New{{$Table}}()
    {{range .Fields}}
    target.{{title .Name}},err = {{readType .Type}}(reader)
    if err != nil {
        return
    }
    {{end}}
    return
}


//Write{{$Table}} write {{$Table}} to output stream -- generate by gsc
func Write{{$Table}}(writer gorpc.Writer,val *{{$Table}}) (err error) {
    {{range .Fields}}
    err = {{writeType .Type}}(writer,val.{{title .Name}})
    if err != nil {
        return
    }
    {{end}}
    return nil
}

{{end}}


{{define "contract"}}{{$Contract := title .Name}}

//{{$Contract}} -- generate by gsc
type {{$Contract}} interface {
    {{range .Methods}}
    {{title .Name}}{{params .Params}}{{returnParam .Return}}
    {{end}}
}

//_{{$Contract}}Maker -- generate by gs2go
type _{{$Contract}}Maker struct {
    id            uint16          // service id
    impl          {{$Contract}}  // service implement
}
// Make{{$Contract}} -- generate by gs2go
func Make{{$Contract}}(id uint16,impl {{$Contract}}) (gorpc.Dispatcher){
    return &_{{$Contract}}Maker{
        id:      id,
        impl:    impl,
    }
}
// Dispatch implement gsrpc.Dispatcher
func (maker *_{{$Contract}}Maker) Dispatch(call *gsrpc.Request) (callReturn *gsrpc.Response, err error) {

    defer func(){
        if e := recover(); e != nil {
            err = gserrors.New(e.(error))
        }
    }()

    switch call.Method {
    {{range .Methods}}{{$Name := title .Name}}
    case {{.ID}}:
        if len(call.Params) != {{.ParamsCount}} {
            err = gserrors.Newf(nil,"{{$Contract}}#{{$Name}} expect {{.ParamsCount}} params but got :%d",len(call.Params))
            return
        }

        {{range .Params}}
        var {{.Name}} {{typeName .Type}}
        {{.Name}},err = {{readType .Type}}(bytes.NewBuffer(call.Params[{{.ID}}].Content))
        if err != nil {
            err = gserrors.Newf(err,"read {{$Contract}}#{{$Name}} arg({{.Name}}) err")
            return
        }
        {{end}}


        {{if notVoid .Return}}
        var retval {{typeName .Return}}
        {{end}}

        {{returnArgs .Return}} = maker.impl.{{$Name}}{{callArgs .Params}}

        if err != nil {

            {{if .Exceptions}}

            var buff bytes.Buffer

            switch err.(type) {
            {{range .Exceptions}}
            case {{typeName .Type}}:

                err = {{writeType .Type}}(&buff,err.({{typeName .Type}}))

                if err != nil {
                    return
                }

            {{end}}
            default:
                return
            }

            callReturn = &gsrpc.Response{
                ID : call.ID,
                Service:call.Service,
            }

            callReturn.Content = buff.Bytes()

            {{end}}

            return
        }

        {{if notVoid .Return}}

        var buff bytes.Buffer

        err = {{writeType .Return}}(&buff,retval)

        if err != nil {
            return
        }

        callReturn = &gsrpc.Response{
            ID : call.ID,
            Service:call.Service,
        }

        callReturn.Content = buff.Bytes()

        {{else}}
        callReturn = &gsrpc.Response{
            ID : call.ID,
            Service:call.Service,
        }
        {{end}}

        return


    {{end}}
    }
    err = gserrors.Newf(nil,"unknown {{$Contract}}#%d method",call.Method)
    return
}

{{end}}


//_{{$Contract}}Binder the remote service proxy binder
type _{{$Contract}}Binder struct {
    id            uint16          // service id
    channel       gsrpc.Channel   // contract bind channel
}
// Bind{{$Contract}} bind remote service and return remote service's proxy object
func Bind{{$Contract}}(id uint16,channel gsrpc.Channel) {{$Contract}} {
    return &_{{$Contract}}Binder{id:id,channel:channel }
}

{{range .Methods}}
{{$Name := symbol .Name}}
//{{$Name}} -- generate by gsc
func (binder *{{$Contract}}Binder){{$Name}}{{params .Params}}{{returnParams .Return}}{
    defer func(){
       if e := recover(); e != nil {
           err = gserrors.New(e.(error))
       }
    }()
    call := &gsrpc.Call{
       Service:uint16(binder.id),
       Method:{{.ID}},
    }
    {{range .Params}}
    var param{{.ID}} bytes.Buffer
    err = {{writeType .Type}}(&param{{.ID}},arg{{.ID}})
    if err != nil {
        return
    }
    call.Params = append(call.Params,&gsrpc.Param{Content:param{{.ID}}.Bytes()})
    {{end}}
    var future gsrpc.Future
    future, err = binder.channel.Send(call)
    if err != nil {
        return
    }
    {{if .Return}}
    var callReturn *gsrpc.Return
    callReturn, err = future.Wait()
    if err != nil {
        return
    }
    {{range .Return}}
    ret{{.ID}},err = {{readType .Type}}(bytes.NewBuffer(callReturn.Params[{{.ID}}].Content))
    if err != nil {
        err = gserrors.Newf(err,"read {{$Contract}}#{{$Name}} return{{.ID}}")
        return
    }
    {{end}}
    {{else}}
    _, err = future.Wait()
    {{end}}
    return
}
{{end}}


{{define "create_array"}}func() {{typeName .}} {

    var buff {{typeName .}}

    {{if builtin .Component}}
    {{else}}
    for i := uint16(0); i < {{.Size}}; i ++ {
        buff[i] = {{defaultVal .Component}}
    }
    {{end}}

    return buff
}(){{end}}



{{define "readList"}}func(reader gorpc.Reader)({{typeName .}},error) {
    length ,err := gorpc.ReadUInt16(reader)
    if err != nil {
        return nil,err
    }
    buff := make({{typeName .}},length)
    for i := uint16(0); i < length; i ++ {
        buff[i] ,err = {{readType .Component}}(reader)
        if err != nil {
            return buff,err
        }
    }
    return buff,nil
}{{end}}


{{define "readByteList"}}func(reader gorpc.Reader)({{typeName .}},error) {
    length ,err := gorpc.ReadUInt16(reader)
    if err != nil {
        return nil,err
    }
    if length == 0 {
        return nil,nil
    }
    buff := make({{typeName .}},length)
    err = gorpc.ReadBytes(reader,buff)
    return buff,err
}{{end}}

{{define "readArray"}}func(reader gorpc.Reader)({{typeName .}},error) {
    var buff {{typeName .}}
    if err != nil {
        return buff,err
    }
    for i := uint16(0); i < {{.Size}}; i ++ {
        buff[i] ,err = {{readType .Component}}(reader)
        if err != nil {
            return buff,err
        }
    }
    return buff,nil
}{{end}}

{{define "readByteArray"}}func(reader gorpc.Reader)({{typeName .}},error) {
    var buff {{typeName .}}
    if err != nil {
        return buff,err
    }
    err = gorpc.ReadBytes(reader,buff[:])
    return buff,err
}{{end}}


{{define "writeList"}}func(writer gorpc.Writer,val {{typeName .}})(error) {
    gorpc.WriteUInt16(writer,uint16(len(val)))
    for _,c:= range val {
        err := {{writeType .Component}}(writer,c)
        if err != nil {
            return err
        }
    }
    return nil
}{{end}}
{{define "writeByteList"}}func(writer gorpc.Writer,val {{typeName .}})(error) {
    err := gorpc.WriteUInt16(writer,uint16(len(val)))
    if err != nil {
        return err
    }
    if len(val) != 0 {
        return gorpc.WriteBytes(writer,val)
    }
    return nil
}{{end}}
{{define "writeArray"}}func(writer gorpc.Writer,val {{typeName .}})(error) {
    for _,c:= range val {
        err := {{writeType .Component}}(writer,c)
        if err != nil {
            return err
        }
    }
    return nil
}{{end}}
{{define "writeByteArray"}}func(writer gorpc.Writer,val {{typeName .}})(error) {
    return gorpc.WriteBytes(writer,val[:])
}{{end}}

`