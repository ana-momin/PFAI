const hosted={chat:"https://pfai-signal-chat.onrender.com",auth:"https://pfai-keyline.vercel.app",crawler:"https://pfai-orbit.vercel.app"};
document.querySelectorAll("[data-app]").forEach(a=>a.href=hosted[a.dataset.app]);

const dot=document.querySelector(".cursor-dot"),ring=document.querySelector(".cursor-ring");
let mouse={x:innerWidth/2,y:innerHeight/2},trail={...mouse};
addEventListener("mousemove",e=>{mouse.x=e.clientX;mouse.y=e.clientY;dot.style.transform=`translate(${e.clientX}px,${e.clientY}px) translate(-50%,-50%)`});
function cursorLoop(){trail.x+=(mouse.x-trail.x)*.14;trail.y+=(mouse.y-trail.y)*.14;ring.style.transform=`translate(${trail.x}px,${trail.y}px) translate(-50%,-50%)`;requestAnimationFrame(cursorLoop)}cursorLoop();
document.querySelectorAll("[data-cursor]").forEach(el=>{el.addEventListener("mouseenter",()=>{document.body.classList.add("cursor-open");ring.querySelector("span").textContent=el.dataset.cursor});el.addEventListener("mouseleave",()=>document.body.classList.remove("cursor-open"))});

document.querySelectorAll(".magnetic").forEach(el=>{el.addEventListener("mousemove",e=>{const r=el.getBoundingClientRect();el.style.transform=`translate(${(e.clientX-r.left-r.width/2)*.15}px,${(e.clientY-r.top-r.height/2)*.15}px)`});el.addEventListener("mouseleave",()=>el.style.transform="")});

const observer=new IntersectionObserver(entries=>entries.forEach(entry=>{if(entry.isIntersecting){entry.target.classList.add("visible");observer.unobserve(entry.target)}}),{threshold:.12});
document.querySelectorAll(".reveal,.reveal-project").forEach(el=>observer.observe(el));

document.querySelectorAll(".project").forEach(card=>{const depth=card.querySelector("[data-depth]");card.addEventListener("mousemove",e=>{const r=card.getBoundingClientRect(),x=(e.clientX-r.left)/r.width-.5,y=(e.clientY-r.top)/r.height-.5;depth.style.transform=`translate(${x*24}px,${y*18}px) rotateX(${-y*3}deg) rotateY(${x*3}deg)`});card.addEventListener("mouseleave",()=>depth.style.transform="")});

const core=document.querySelector(".core"),progress=document.querySelector(".progress");
addEventListener("scroll",()=>{const max=document.documentElement.scrollHeight-innerHeight,ratio=scrollY/max;progress.style.width=`${ratio*100}%`;if(innerWidth>900)core.style.transform=`translateY(calc(-50% + ${scrollY*.06}px)) rotate(${scrollY*.025}deg)`},{passive:true});
