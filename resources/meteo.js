// Manage meteo (reload every hour to refresh info)
var Meteo = {
    scriptRef:null,
    autoInit:function(){
        this.init();
        return this;
    },
    init:function(){
        var _self = this;
        TimeEventManager.add(PatternManager.everyHours(),function(){_self.loadMeteo();},true);
    },
    // load meteo script. Remove existing if necessary
    loadMeteo:function(){
        // remove script
        if(this.scriptRef != null){
            this.scriptRef.remove();
        }
        // Remove iframe before
        $('iframe','#widget_e51c7a153a96e46f31cafc3b97dc885c').remove();

        this.scriptRef = document.createElement("script");
        this.scriptRef.type = "text/javascript";
        this.scriptRef.async = true;
        this.scriptRef.src = "http://services.my-meteo.fr/widget/js3.php?ville=251&format=petit-horizontal&nb_jours=3&icones&horaires&c1=ffffff&c2=ffffff&c3=000000&c4=000000&c5=00d2ff&c6=ffc334&police=0&t_icones=1&x=422&y=56&id=e51c7a153a96e46f31cafc3b97dc885c";
        var z = document.getElementsByTagName("script")[0];
        z.parentNode.insertBefore(this.scriptRef, z);
    }
 }.autoInit();