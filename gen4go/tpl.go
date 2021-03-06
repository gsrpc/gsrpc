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
    return {{if enumSize . | eq 4}} gorpc.WriteUInt32(writer,uint32(val)) {{else}} gorpc.WriteByte(writer,byte(val)) {{end}}
}

//Read{{$Enum}} write enum to output stream
func Read{{$Enum}}(reader gorpc.Reader)({{$Enum}}, error){
    val,err := {{if enumSize . | eq 4}} gorpc.ReadUInt32(reader) {{else}} gorpc.ReadByte(reader) {{end}}
    return {{$Enum}}(val),err
}

//String implement Stringer interface
func (val {{$Enum}}) String() string {
    switch val {
        {{range .Constants}}
        case {{.Value}}:
            return "{{$Enum}}.{{title .Name}}"
        {{end}}
    }
    return fmt.Sprintf("enum(Unknown(%d))",val)
}

{{end}}

{{define "table"}} {{$Table := title .Name}}

//{{$Table}} -- generate by gsc
type {{$Table}} struct {
    {{range .Fields}}
    {{title .Name}} {{typeName .Type}}
    {{end}}
}

{{if isException .}}
//Error implement error interface
func (e * {{$Table}}) Error() string {
    return "{{$Table}} error"
}
{{end}}

//New{{$Table}} create new struct object with default field val -- generate by gsc
func New{{$Table}}() *{{$Table}} {
    return &{{$Table}}{
        {{range .Fields}}
        {{title .Name}}: {{defaultVal .Type}},
        {{end}}
    }
}

{{if isPOD .}}
//Read{{$Table}} read {{$Table}} from input stream -- generate by gsc
func Read{{$Table}}(reader gorpc.Reader) (target *{{$Table}},err error) {
    target = New{{$Table}}()

    {{range .Fields}}

    {
        target.{{title .Name}},err = {{readType .Type}}(reader)

        if err != nil {
            return
        }
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

{{else}}
//Read{{$Table}} read {{$Table}} from input stream -- generate by gsc
func Read{{$Table}}(reader gorpc.Reader) (target *{{$Table}},err error) {
    target = New{{$Table}}()

    var fields byte

    fields,err = gorpc.ReadByte(reader)

    if err != nil {
        return
    }

    {{range .Fields}}

    {
        var tag byte
        tag,err = gorpc.ReadByte(reader)

        if err != nil {
            return
        }

        if tag != byte(gorpc.TagSkip) {
            target.{{title .Name}},err = {{readType .Type}}(reader)

            if err != nil {
                return
            }
        }

        fields --

        if fields == 0 {
            return
        }
    }
    {{end}}

    for i :=0;i < int(fields); i ++ {

        var tag byte

        tag,err = gorpc.ReadByte(reader)

        if err != nil {
            return
        }

        if tag == byte(gorpc.TagSkip) {
            continue
        }

        gorpc.SkipRead(reader,gorpc.Tag(tag))
    }


    return
}


//Write{{$Table}} write {{$Table}} to output stream -- generate by gsc
func Write{{$Table}}(writer gorpc.Writer,val *{{$Table}}) (err error) {

    err = gorpc.WriteByte(writer,byte({{len .Fields}}))

    if err != nil {
        return
    }

    {{range .Fields}}
    gorpc.WriteByte(writer,byte({{tagValue .Type}}))
    err = {{writeType .Type}}(writer,val.{{title .Name}})
    if err != nil {
        return
    }
    {{end}}
    return nil
}
{{end}}


{{end}}


{{define "contract"}}{{$Contract := title .Name}}

//{{$Contract}} -- generate by gsc
type {{$Contract}} interface {
    {{range .Methods}}
    {{title .Name}}{{params .Params}}{{returnParam .Return}}
    {{end}}
}

const (
    NameOf{{$Contract}} = "{{.FullName}}"
)

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
// ID implement gorpc.Dispatcher
func (maker *_{{$Contract}}Maker) ID() uint16 {
    return maker.id
}

// ID implement gorpc.Dispatcher
func (maker *_{{$Contract}}Maker) String() string {
    return "{{.FullName}}"
}

// Dispatch implement gorpc.Dispatcher
func (maker *_{{$Contract}}Maker) Dispatch(call *gorpc.Request) (callReturn *gorpc.Response, err error) {

    defer func(){
        if e := recover(); e != nil {
            err = gserrors.New(e.(error))
        }
    }()

    traceflag := trace.Flag()

    if traceflag {
        traceRPC := trace.RPC(call.Trace,uint32(call.Service) << 16 | uint32(call.Method),call.Prev)

        traceRPC.Start()

        defer traceRPC.End()
    }

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

        callSite := &gorpc.CallSite {
            ID : uint32(call.Service) << 16 | uint32({{.ID}}),
            Trace : call.Trace,
        }


        {{if isAsync . | not }}{{if notVoid .Return}}
        var retval {{typeName .Return}}
        {{end}}{{end}}

        {{returnArgs .Return}} = maker.impl.{{$Name}}{{callArgs .Params}}

        {{if isAsync . | not }}
        if err != nil {

            {{if .Exceptions}}

            var buff bytes.Buffer

            id := int8(-1)

            switch err.(type) {
            {{range .Exceptions}}
            case {{typeName .Type}}:

                err = {{writeType .Type}}(&buff,err.({{typeName .Type}}))

                if err != nil {
                    return
                }

                id = {{.ID}}

            {{end}}
            default:
                return
            }

            callReturn = &gorpc.Response{
                ID : call.ID,
                Exception:id,
                Trace:call.Trace,
            }

            callReturn.Content = buff.Bytes()

            err = nil

            {{end}}

            return
        }

        {{if notVoid .Return}}

        var buff bytes.Buffer

        err = {{writeType .Return}}(&buff,retval)

        if err != nil {
            return
        }

        callReturn = &gorpc.Response{
            ID : call.ID,
            Exception:int8(-1),
            Trace:call.Trace,
        }

        callReturn.Content = buff.Bytes()

        {{else}}
        callReturn = &gorpc.Response{
            ID : call.ID,
            Exception:int8(-1),
            Trace:call.Trace,
        }
        {{end}}
        {{end}}

        return


    {{end}}
    }
    err = gserrors.Newf(nil,"unknown {{$Contract}}#%d method",call.Method)
    return
}


//_{{$Contract}}Binder the remote service proxy binder
type _{{$Contract}}Binder struct {
    id            uint16          // service id
    channel       gorpc.Channel   // contract bind channel
}
// Bind{{$Contract}} bind remote service and return remote service's proxy object
func Bind{{$Contract}}(id uint16,channel gorpc.Channel) {{$Contract}} {
    return &_{{$Contract}}Binder{id:id,channel:channel }
}

{{range .Methods}}
{{$Name := title .Name}}
//{{$Name}} -- generate by gsc
func (binder *_{{$Contract}}Binder){{$Name}}{{params .Params}}{{returnParam .Return}}{
    defer func(){
       if e := recover(); e != nil {
           err = gserrors.New(e.(error))
       }
    }()

    var traceID uint64
    var traceParentID uint32

    if trace.Flag() {
        if callSite != nil  {
            traceID = callSite.Trace
            traceParentID = callSite.ID
        } else {
            traceID = trace.NewTrace()
        }

        traceRPC := trace.RPC(traceID,uint32(binder.id) << 16 | uint32({{.ID}}),traceParentID)

        traceRPC.Start()

        defer traceRPC.End()
    }

    call := &gorpc.Request{
       Service:uint16(binder.id),
       Method:{{.ID}},
       Trace:traceID,
       Prev:traceParentID,
    }


    {{range .Params}}
    var param{{.ID}} bytes.Buffer
    err = {{writeType .Type}}(&param{{.ID}},{{.Name}})
    if err != nil {
        return
    }
    call.Params = append(call.Params,&gorpc.Param{Content:param{{.ID}}.Bytes()})
    {{end}}

    {{if isAsync .}}
    err = binder.channel.Post(call)
    return
    {{else}}
    var future gorpc.Future
    future, err = binder.channel.Send(call)
    if err != nil {
        return
    }

    var callReturn *gorpc.Response
    callReturn, err = future.Wait()
    if err != nil {
        return
    }

    //TODO: handler response callsite

    if callReturn.Exception != -1 {
        switch callReturn.Exception {
        {{range .Exceptions}}
        case {{.ID}}:
            var exception error
            exception,err = {{readType .Type}}(bytes.NewBuffer(callReturn.Content))

            if err != nil {
                err = gserrors.Newf(err,"read {{$Contract}}#{{$Name}} return")
            } else {
                err = exception
            }

            return
        {{end}}
        default:
            err = gserrors.Newf(gorpc.ErrRPC,"catch unknown exception(%d) for {{$Contract}}#{{$Name}}",callReturn.Exception)
            return
        }
    }



    {{if notVoid .Return}}
    retval,err = {{readType .Return}}(bytes.NewBuffer(callReturn.Content))

    if err != nil {
        err = gserrors.Newf(err,"read {{$Contract}}#{{$Name}} return")
        return
    }
    {{end}}
    {{end}}
    return
}
{{end}}

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

    length ,err := gorpc.ReadUInt16(reader)

    if err != nil {
        return buff,err
    }

    if length != {{.Size}} {
        return buff,gserrors.Newf(nil,"check array size failed")
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

    length ,err := gorpc.ReadUInt16(reader)
    if err != nil {
        return buff,err
    }

    if length != {{.Size}} {
        return buff,gserrors.Newf(nil,"check array size failed")
    }

    if length == 0 {
        return buff,nil
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
    gorpc.WriteUInt16(writer,uint16(len(val)))
    for _,c:= range val {
        err := {{writeType .Component}}(writer,c)
        if err != nil {
            return err
        }
    }
    return nil
}{{end}}

{{define "writeByteArray"}}func(writer gorpc.Writer,val {{typeName .}})(error) {
    err := gorpc.WriteUInt16(writer,uint16(len(val)))
    if err != nil {
        return err
    }
    if len(val) != 0 {
        return gorpc.WriteBytes(writer,val[:])
    }
    return nil
}{{end}}

`
