(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([[32],{ef36:function(t,e,n){"use strict";n.r(e);var r=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",[n("div",{staticClass:"row justify-center"},[n("div",{staticClass:"row q-mt-lg",staticStyle:{width:"600px"}},[n("div",{staticClass:"col-6 text-h6 text-purple"},[t._v(" 快速进入某道题的所有讨论")]),n("Input",{staticClass:"col-6",attrs:{search:"","enter-button":"",placeholder:"输入题目索引 如P1000..."},on:{"on-search":t.search},model:{value:t.index,callback:function(e){t.index=e},expression:"index"}})],1)]),n("div",{staticClass:"q-pa-md row justify-center q-mt-lg"},[n("GlobalPostCard",{attrs:{post_list:t.discuss_list},on:{getPosts:t.get_posts,enter:t.enter}})],1)])},s=[],i=(n("23bf"),n("06db"),n("7f7f"),n("8a81"),n("1c4c"),n("5df3"),n("cadf"),n("ac6a"),n("dc9a"));function o(t,e){var n="undefined"!==typeof Symbol&&t[Symbol.iterator]||t["@@iterator"];if(!n){if(Array.isArray(t)||(n=a(t))||e&&t&&"number"===typeof t.length){n&&(t=n);var r=0,s=function(){};return{s:s,n:function(){return r>=t.length?{done:!0}:{done:!1,value:t[r++]}},e:function(t){throw t},f:s}}throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}var i,o=!0,c=!1;return{s:function(){n=n.call(t)},n:function(){var t=n.next();return o=t.done,t},e:function(t){c=!0,i=t},f:function(){try{o||null==n.return||n.return()}finally{if(c)throw i}}}}function a(t,e){if(t){if("string"===typeof t)return c(t,e);var n=Object.prototype.toString.call(t).slice(8,-1);return"Object"===n&&t.constructor&&(n=t.constructor.name),"Map"===n||"Set"===n?Array.from(t):"Arguments"===n||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)?c(t,e):void 0}}function c(t,e){(null==e||e>t.length)&&(e=t.length);for(var n=0,r=new Array(e);n<e;n++)r[n]=t[n];return r}var u={name:"discussIndex",components:{GlobalPostCard:i["a"]},data:function(){return{index:"",discuss_list:[]}},methods:{search:function(){var t=this.index.trim();""!=t?("p"==t[0]?t="P"+t.substr(1):"P"!=t[0]&&(t="P"+t),this.$router.push({name:"discussProblem",params:{index:t}})):this.$notify("error","请输入题目索引")},get_posts:function(t,e){var n=this;this.$req.get("getPostList",{kind:"puzzle",l:t,r:e}).then((function(t){if(void 0==t.errno){n.discuss_list=t.data;var e,r=o(n.discuss_list);try{for(r.s();!(e=r.n()).done;){var s=e.value;""!=s.tags?s.tags=JSON.parse(s.tags):s.tags=[]}}catch(i){r.e(i)}finally{r.f()}}}))},enter:function(t){this.$router.push({name:"discussion",params:{puzzle_id:t.post_id}})}},mounted:function(){setTimeout(this.get_posts(1,10),500)}},l=u,d=n("4023"),f=Object(d["a"])(l,r,s,!1,null,"4de75383",null);e["default"]=f.exports}}]);