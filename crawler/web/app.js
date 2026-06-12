const $=s=>document.querySelector(s);
["limit","workers"].forEach(x=>$("#"+x).oninput=e=>$("#"+x+"-val").textContent=e.target.value);

$("#crawl").onsubmit=async e=>{
  e.preventDefault();
  const button=e.currentTarget.querySelector("button"),url=$("#url").value;
  button.disabled=true;
  button.textContent="Workers are crawling...";
  $("#mission").hidden=false;
  $("#results").innerHTML="";
  $("#report-count").textContent="0 pages";
  ["pages","links"].forEach(x=>$("#"+x).textContent="0");
  $("#success").textContent="0%";
  $("#speed").textContent="0ms";
  $("#target").textContent=url;
  $("#state span").innerHTML="Crawling<small>Go workers are exploring the site</small>";
  $("#mission").scrollIntoView({behavior:"smooth"});
  try{
    const r=await fetch("/api/crawl",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({url,maxPages:+$("#limit").value,concurrency:+$("#workers").value})});
    if(!r.ok)throw Error(await r.text());
    const data=await r.json();
    show(data.pages);
  }catch(err){
    $("#state span").innerHTML="Crawl stopped<small>Please review the destination</small>";
    $("#error").textContent=err.message;
    $("#error").style.display="block";
    setTimeout(()=>$("#error").style.display="none",3500);
  }finally{
    button.disabled=false;
    button.innerHTML="Start crawling <span>→</span>";
  }
};

function show(items){
  $("#state span").innerHTML="Crawl complete<small>All Go workers returned safely</small>";
  if(!items.length){$("#results").innerHTML='<div class="empty">No pages were returned for this destination.</div>';return}
  let i=0,t=setInterval(()=>{
    if(i>=items.length){clearInterval(t);return}
    const r=items[i++],el=document.createElement("div");
    el.className="result "+(r.status>=400||r.error?"bad":"");
    el.innerHTML=`<span class="code">${r.status||"ERR"}</span><div class="resource"><b>${esc(r.title||r.error||"Unreachable")}</b><small>${esc(r.url)}</small></div><span>${r.links}</span><span>${r.durationMs}ms</span>`;
    $("#results").append(el);
    $("#report-count").textContent=i+(i===1?" page":" pages");
    metrics(items.slice(0,i));
  },55);
}

function metrics(a){
  $("#pages").textContent=a.length;
  $("#success").textContent=Math.round(a.filter(x=>x.status<400&&x.status>0).length/a.length*100)+"%";
  $("#speed").textContent=Math.round(a.reduce((n,x)=>n+x.durationMs,0)/a.length)+"ms";
  $("#links").textContent=a.reduce((n,x)=>n+x.links,0);
}
function esc(s){const d=document.createElement("div");d.textContent=s;return d.innerHTML}
