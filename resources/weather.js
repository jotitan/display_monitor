// get weather on api wunderground

var Weather = {
    autoInit:function(){
        this.init();
        return this;
    },
    init:function(){
        var _self = this;
        TimeEventManager.add(PatternManager.everyMinute(),function(){_self.load();},true);
    },
    load:function(){
        var url = 'http://api.wunderground.com/api/59cf54a22178a143/conditions/q/CA/pws:IVAIRESS3.json';
        $.ajax({
           url:url,
           dataType:'json',
           success:function(data){
                $('.local-temp').html(data.current_observation.temp_c);
           }
        });
    }
}.autoInit();

