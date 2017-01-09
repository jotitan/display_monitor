/**
 * Created by Titan on 06/12/2016.
 */

var MUSIC_URL = 'http://localhost:9009';


// Share manage
var Share = {
    getShares: function (url, callback) {
        $.ajax({
            url: url + '/shares',
            dataType: 'json',
            success: function (data) {
                callback(data);
            },error:function(){
                callback(null);
            }
        });
    },
    createRemote:function(id,url,target){
        var manager = {id:id};
        var sse = new EventSource(url + '/share?id=' + id + '&device=screen-displayer');

        sse.addEventListener('close',function(data){
            manager.disable();
        });
        sse.addEventListener('load',function(response){
            var id = response.data;
            $.ajax({
                url:url + '/musicInfo?id=' + id,
                success:function(data){
                    if(target!=null){
                        target.updateMusic(JSON.parse(data));
                    }
                }
            })
        });

        manager.sse = sse;
        manager.event = function(event,data){
            data = data == null ? "" : data;
            $.ajax({
                url:url + '/shareUpdate',
                data:{id:this.id,event:event,data:data}
            });
        };
        return manager;
    }
}

var RemoteControlManager = {
    manager:null,
    div:null,
    divSelect:null,
    url:'',
    init:function(idDiv,idSelect,musicUrl){
        var _self = this;
        this.url = musicUrl || '';
        this.div = $(idDiv);
        this.divSelect = $(idSelect);
        $('.play',this.div).bind('click',function(){
            _self.manager.event('play');
            $('.play',this.div).hide();
            $('.pause',this.div).show();
        });

        $('.pause',this.div).bind('click',function(){
            _self.manager.event('pause');
            $('.pause',this.div).hide();
            $('.play',this.div).show();
        });

        $('.previous',this.div).bind('click',function() {
            _self.manager.event('previous');
        });

        $('.next',this.div).bind('click',function() {
            _self.manager.event('next');
        });

        this.divSelect.bind('change',function(){
            _self.manager = Share.createRemote($(this).val(),_self.url,_self);
            _self.divSelect.hide();
            _self.div.show();
        });
        Share.getShares(this.url,function(data) {
            if(data == null || data.length == 0){
                _self.divSelect.hide();
                return;
            }
            _self.divSelect.empty().append('<option>...</option>');
            data.forEach(function (s) {
                _self.divSelect.append('<option value="' + s.Id + '">' + s.Name + '</option>');
            });
        });
        return this;
    },
    updateMusic:function(music) {
        $('.title',this.div).html(music.title + " - " + music.artist);
    }
}.init('.musicRemoteControl > .remote','.musicRemoteControl > .select-share',MUSIC_URL);