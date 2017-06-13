/* Display image */



// Manage rolling image in background
var Gallery = {
    url:'/image?format=jpg',
    imgA:null,
    imgB:null,
    autoInit:function(){
        if($('.gallery').length > 0){
            this.imgA = $('img:first','.gallery')
            this.imgB = $('img:last','.gallery')
            this.init();
        }
        return this;
    },
    init:function(){
        var _self = this;
        $('.gallery').bind('click',toggleFullScreen);
        $('.folderName').bind('click',function(){Gallery.changeFolder();})
        TimeEventManager.add(PatternManager.every30Seconds(),function(){_self.setImage();},true);
    },
    setImage:function(){
        if(this.imgA.is(':visible')){
            this._loadAndSwitch(this.imgB,this.imgA);
        }else{
            this._loadAndSwitch(this.imgA,this.imgB);
        }
    },
    changeFolder:function(){
        $.ajax({url:'/change',success:function(){Gallery.loadFolderName();}});
    },
    changeGalleryType:function(){
      $.ajax({url:'/changeGallery',success:function(){Gallery.loadFolderName();}});
    },
    loadFolderName:function(){
        $.ajax({url:'/getFolderName',success:function(data){$('.folderName').html(data)}});
    },
    _loadAndSwitch:function(imgHidden,imgVisible){
        imgHidden.attr('src',this.url + "&rand=" + Math.random());
        imgHidden.unbind('load').bind('load',function(){
            Gallery.loadFolderName();
            if(imgHidden.width() > imgHidden.height()){
                imgHidden.css('width','100%').css('height','');
            }else{
                imgHidden.css('height','100%').css('width','');
            }
            imgHidden.animate({opacity:1},1500);
            imgVisible.animate({opacity:0},1500);
        });
    }
}.autoInit();
