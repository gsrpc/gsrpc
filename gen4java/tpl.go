package gen4java

var t4java = `
{{define "enum"}}{{$Enum := title .Name}}
/*
 * {{title .Name}} generate by gs2java,don't modify it manually
 */
public enum {{title .Name}} {
    {{enumFields .}}
    private {{enumType .}} value;
    {{title .Name}}({{enumType .}} val){
        this.value = val;
    }
    @Override
    public String toString() {
        switch(this.value)
        {
        {{range .Constants}}
        case {{.Value}}:
            return "{{title .Name}}";
        {{end}}
        }
        return String.format("{{title .Name}}#%d",this.value);
    }
    public {{enumType .}} getValue() {
        return this.value;
    }
    public void Marshal(Writer writer) throws Exception
    {
        {{if enumSize . | eq 4}} writer.WriteUInt32(getValue()); {{else}} writer.WriteByte(getValue()); {{end}}
    }
    public static {{title .Name}} Unmarshal(Reader reader) throws Exception
    {
        {{enumType .}} code =  {{if enumSize . | eq 4}} reader.ReadUInt32(); {{else}} reader.ReadByte(); {{end}}
        switch(code)
        {
        {{range .Constants}}
        case {{.Value}}:
            return {{$Enum}}.{{title .Name}};
        {{end}}
        }
        throw new Exception("unknown enum constant :" + code);
    }
}
{{end}}

{{define "exception"}}{{$Struct := title .Name}}
public class {{$Struct}} extends Exception
{
{{range .Fields}}
    private  {{typeName .Type}} {{fieldName .Name}} = {{defaultVal .Type}};
{{end}}

{{range .Fields}}
    public {{typeName .Type}} get{{title .Name}}()
    {
        return this.{{fieldName .Name}};
    }
    public void set{{title .Name}}({{typeName .Type}} arg)
    {
        this.{{fieldName .Name}} = arg;
    }
{{end}}
    public void Marshal(Writer writer)  throws Exception
    {
{{range .Fields}}
        {{marshalField .}}
{{end}}
    }
    public void Unmarshal(Reader reader) throws Exception
    {
{{range .Fields}}
        {{unmarshalField .}}
{{end}}
    }
}
{{end}}


{{define "table"}}{{$Struct := title .Name}}
/*
 * {{title .Name}} generate by gs2java,don't modify it manually
 */
public class {{$Struct}}
{
{{range .Fields}}
    private  {{typeName .Type}} {{fieldName .Name}} = {{defaultVal .Type}};
{{end}}

{{range .Fields}}
    public {{typeName .Type}} get{{title .Name}}()
    {
        return this.{{fieldName .Name}};
    }
    public void set{{title .Name}}({{typeName .Type}} arg)
    {
        this.{{fieldName .Name}} = arg;
    }
{{end}}
    public void Marshal(Writer writer)  throws Exception
    {
{{range .Fields}}
        {{marshalField .}}
{{end}}
    }
    public void Unmarshal(Reader reader) throws Exception
    {
{{range .Fields}}
        {{unmarshalField .}}
{{end}}
    }
}
{{end}}


{{define "contract"}}{{$Contract := title .Name}}

public interface {{$Contract}} {
{{range .Methods}}
    {{returnParam .Return}} {{title .Name}} {{params .Params}} throws Exception;
{{end}}
}

{{end}}


{{define "dispatcher"}}
{{$Contract := title .Name}}
/*
 * {{title .Name}} generate by gs2java,don't modify it manually
 */
public final class {{$Contract}}Dispatcher implements com.gsrpc.Dispatcher {

    private {{$Contract}} service;

    public {{$Contract}}Dispatcher({{$Contract}} service) {
        this.service = service;
    }

    public com.gsrpc.Response Dispatch(com.gsrpc.Request call) throws Exception
    {
        switch(call.getMethod()){
        {{range .Methods}}
        case {{.ID}}: {
{{range .Params}}{{unmarshalParam . "call" 4}}{{end}}

                try{
                    {{methodcall .}}

                    com.gsrpc.Response callReturn = new com.gsrpc.Response();
                    callReturn.setID(call.getID());
                    callReturn.setService(call.getService());
                    callReturn.setException((byte)-1);

                    {{if notVoid .Return}}
    {{marshalReturn .Return "ret" 4}}
                    callReturn.setContent(returnParam);
                    {{end}}

                    return callReturn;

                } catch(Exception e){
                    {{range .Exceptions}}
                    if(e instanceof {{typeName .Type}}){

                        com.gsrpc.BufferWriter writer = new com.gsrpc.BufferWriter();

                        (({{typeName .Type}})e).Marshal(writer);

                        com.gsrpc.Response callReturn = new com.gsrpc.Response();
                        callReturn.setID(call.getID());
                        callReturn.setService(call.getService());
                        callReturn.setException((byte){{.ID}});
                        callReturn.setContent(writer.Content());

                        return callReturn;
                    }
                    {{end}}
                }
            }
        {{end}}
        }
        return null;
    }
}
{{end}}


{{define "rpc"}}{{$Contract := title .Name}}
/*
 * {{title .Name}} generate by gs2java,don't modify it manually
 */
public final class {{$Contract}}RPC {

    /**
     * gsrpc net interface
     */
    private com.gsrpc.Channel net;

    /**
     * remote service id
     */
    private short serviceID;

    public {{$Contract}}RPC(com.gsrpc.Channel net, short serviceID){
        this.net = net;
        this.serviceID = serviceID;
    }

    {{range .Methods}}{{$Name := title .Name}}
    public {{methodRPC .}} throws Exception {

        com.gsrpc.Request request = new com.gsrpc.Request();

        request.setService(this.serviceID);

        request.setMethod((short){{.ID}});

        {{if .Params}}
        com.gsrpc.Param[] params = new com.gsrpc.Param[{{len .Params}}];
{{marshalParams .Params}}
        request.setParams(params);
        {{end}}

        this.net.send(request,new com.gsrpc.Callback(){
            @Override
            public int getTimeout() {
                return timeout;
            }
            @Override
            public void Return(Exception e,com.gsrpc.Response callReturn){
                if (e != null) {
                    promise.Notify(e,null);
                    return;
                }

                try{

                    if(callReturn.getException() != (byte)-1) {
                        switch(callReturn.getException()) {
                        {{range .Exceptions}}
                        case {{.ID}}:{
                            com.gsrpc.BufferReader reader = new com.gsrpc.BufferReader(callReturn.getContent());

                            {{typeName .Type}} exception = {{defaultVal .Type}};

                            {{readType "exception" .Type 4}}

                            promise.Notify(exception,null);

                            return;
                        }
                        {{end}}
                        default:
                            promise.Notify(new com.gsrpc.RemoteException(String.format("catch unknown exception(%d) for {{$Contract}}#{{$Name}}",callReturn.getException())),null);
                            return;
                        }
                    }

                    {{if notVoid .Return}}
                    {{unmarshalReturn .Return "callReturn" 5}}
                    promise.Notify(null,returnParam);
                    {{else}}
                    promise.Notify(null,null);
                    {{end}}
                }catch(Exception e1) {
                    promise.Notify(e1,null);
                }
            }
        });
    }
    {{end}}
}
{{end}}
`
