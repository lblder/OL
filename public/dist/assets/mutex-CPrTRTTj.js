function r(){let e=!1,t=[];return{acquire:()=>new Promise(n=>{e?t.push(n):(e=!0,n())}),release:()=>{t.length>0?t.shift()():e=!1}}}export{r as c};
